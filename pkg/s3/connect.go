package s3

import (
	"bytes"
	"context"
	"fmt"
	"imed/internal/config"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Service struct {
	Client *minio.Client
	Cnf    *config.Config
}

func NewS3(cnf *config.Config) (*S3Service, error) {
	client, err := minio.New(cnf.S3_ENDPOINT, &minio.Options{
		Creds:  credentials.NewStaticV4(cnf.S3_ACCESS_KEY, cnf.S3_SECRET_KEY, ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	// Bucket mavjudligini tekshirish va yaratish
	exists, err := client.BucketExists(context.Background(), cnf.S3_BUCKET_NAME)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(context.Background(), cnf.S3_BUCKET_NAME, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	// Bucketni hamma uchun ochiq (Public) qilish
	// ARN formatini to'g'irlaymiz (Contabo uchun identifikator bilan)
	resourceARN := fmt.Sprintf("arn:aws:s3:::%s/*", cnf.S3_BUCKET_NAME)
	if strings.Contains(cnf.S3_BUCKET_NAME, ":") {
		parts := strings.Split(cnf.S3_BUCKET_NAME, ":")
		if len(parts) == 2 {
			resourceARN = fmt.Sprintf("arn:aws:s3::%s:%s/*", parts[0], parts[1])
		}
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "PublicRead",
				"Effect": "Allow",
				"Principal": "*",
				"Action": ["s3:GetObject"],
				"Resource": ["%s"]
			}
		]
	}`, resourceARN)

	err = client.SetBucketPolicy(context.Background(), cnf.S3_BUCKET_NAME, policy)
	if err != nil {
		slog.Error("Failed to set bucket policy", "err", err)
	} else {
		slog.Info("Successfully set S3 bucket policy to Public", "bucket", cnf.S3_BUCKET_NAME)
	}

	return &S3Service{
		Client: client,
		Cnf:    cnf,
	}, nil
}

func (s *S3Service) Upload(fileName, filePath string) (string, error) {
	ext := filepath.Ext(fileName)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.Client.FPutObject(
		context.Background(),
		s.Cnf.S3_BUCKET_NAME,
		fileName,
		filePath,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", s.Cnf.S3_DOMAIN, s.Cnf.S3_BUCKET_NAME, fileName), nil
}

func (s *S3Service) UploadToS3(filePath, videoID string) (string, error) {
	objectName := "youtube/" + videoID + ".mp4"
	return s.Upload(objectName, filePath)
}

func (s *S3Service) UploadFromYouTube(youtubeURL string) (string, string, error) {
	slog.Info("Starting YouTube upload process", "url", youtubeURL)

	// 1. extract video ID
	cleanID, fullID, err := extractVideoID(youtubeURL)
	if err != nil {
		slog.Error("Failed to extract video ID", "err", err)
		return "", "", err
	}
	slog.Info("Extracted video ID", "cleanID", cleanID, "fullID", fullID)

	// 2. download video
	filePath, cleanup, err := downloadYouTube(youtubeURL, cleanID)
	if err != nil {
		slog.Error("Failed to download video from YouTube", "err", err)
		return "", "", err
	}
	if cleanup != nil {
		defer cleanup()
	}
	st, statErr := os.Stat(filePath)
	if statErr != nil {
		return "", "", fmt.Errorf("youtube downloaded file yo‘q yoki ochilmadi: %w (%s)", statErr, filePath)
	}
	if st.Size() < 2048 {
		return "", "", fmt.Errorf("youtube yuklangan fayl juda kichik (size=%d), ehtimol xato", st.Size())
	}
	slog.Info("Video downloaded successfully", "path", filePath, "bytes", st.Size())

	// 3. upload to S3 (Contabo)
	url, err := s.UploadToS3(filePath, cleanID)
	if err != nil {
		slog.Error("Failed to upload video to S3", "err", err, "path", filePath)
		return "", "", err
	}
	slog.Info("Video uploaded to S3 successfully", "url", url)

	return url, fullID, nil
}

func resolveYtDlpCookiePaths() []string {
	if p := strings.TrimSpace(os.Getenv("YT_DLP_COOKIES_FILE")); p != "" {
		return []string{p}
	}
	wd, _ := os.Getwd()
	return []string{
		"internal/secrets/cookies.txt",
		filepath.Join(wd, "internal/secrets/cookies.txt"),
	}
}

func firstExistingCookiesPath() string {
	for _, p := range resolveYtDlpCookiePaths() {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			absPath, absErr := filepath.Abs(p)
			if absErr == nil {
				return absPath
			}
			return p
		}
	}
	return ""
}

func summarizeCmdOutput(output []byte) string {
	const max = 4000
	out := strings.TrimSpace(string(output))
	if len(out) <= max {
		return out
	}
	return "… " + out[len(out)-max:]
}

func ytDlpFormatSelector() string {
	// FFmpeg merge yoqiladi; muxlap mp4 uchun audio/video juftligi bo‘lsa yetarli
	return `bv*[ext=mp4]+ba/bv*+ba/bestvideo[ext=mp4]+bestaudio/bestvideo+bestaudio/b`
}

func buildYtDlpArgs(tmpDir string, youtubePageURL string, extractorArgs string, useBrowserHints bool, cookiesPath string) []string {
	outTemplate := filepath.Join(tmpDir, "merged.%(ext)s")

	args := []string{
		"-f", ytDlpFormatSelector(),
		"-o", outTemplate,
		"--merge-output-format", "mp4",
		"--remux-video", "mp4",
		"--no-playlist",
		"--retries", "5",
		"--fragment-retries", "10",
		"--socket-timeout", "45",
		"--no-check-certificates",
	}

	if strings.TrimSpace(extractorArgs) != "" {
		args = append(args, "--extractor-args", extractorArgs)
	}

	if useBrowserHints {
		args = append(args,
			"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			"--referer", "https://www.youtube.com/",
		)
	}

	if strings.TrimSpace(cookiesPath) != "" {
		args = append(args, "--cookies", cookiesPath)
	}

	args = append(args, youtubePageURL)
	return args
}

func pickMergedOutput(tmpDir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(tmpDir, "merged.*"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("merged.* fayli topilmadi (%s ichida)", tmpDir)
	}
	prefer := func(path string) int {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".mp4":
			return 0
		case ".mkv":
			return 1
		case ".webm":
			return 2
		default:
			return 99
		}
	}

	best := matches[0]
	bestRank := prefer(best)
	for _, m := range matches[1:] {
		r := prefer(m)
		if r < bestRank {
			bestRank = r
			best = m
		}
	}

	fi, statErr := os.Stat(best)
	if statErr != nil {
		return "", fmt.Errorf("merged fayl stat yo‘q: %s (%w)", best, statErr)
	}
	if fi.Size() < 2048 {
		return "", fmt.Errorf("merged fayl juda kichik: %s (size=%d)", best, fi.Size())
	}
	return best, nil
}

func wipeYtDlpTempArtifacts(dir string) {
	patterns := []string{"merged.*", "*.part", "*.frag", "*.ytdl", "*.temp"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		for _, p := range matches {
			if p == "." || p == dir {
				continue
			}
			_ = os.RemoveAll(p)
		}
	}
}

// downloadYouTube merges output to merged.* temporary file and returns absolute path plus cleanup dir.
func downloadYouTube(youtubePageURL, videoID string) (absPath string, cleanup func(), err error) {
	tmpDir, err := os.MkdirTemp("", "imed-yt-*")
	if err != nil {
		return "", nil, err
	}
	cleanupRan := false
	cleanupFn := func() {
		if cleanupRan {
			return
		}
		cleanupRan = true
		_ = os.RemoveAll(tmpDir)
	}

	// extractor urinishlari: YouTube CDN o‘zgarganda turli player_client lar boshqacha ishlaydi
	attempts := []struct {
		extractor string
		useHints  bool
	}{
		{`youtube:player_client=tv_embedded`, true},
		{`youtube:player_client=android_creator,android`, true},
		{`youtube:player_client=mweb`, true},
		{`youtube:player_client=ios,tv_android`, false},
		{`youtube:player_client=web`, false},
		{"", false},
	}

	cookiesPath := firstExistingCookiesPath()
	if cookiesPath != "" {
		slog.Info("yt-dlp cookies", "path", cookiesPath)
	} else {
		slog.Warn("yt-dlp: cookie fayl yo‘q; qiyin videolar uchun YT_DLP_COOKIES_FILE yoki internal/secrets/cookies.txt qo‘shing")
	}

	wipeYtDlpTempArtifacts(tmpDir)

	var lastErr error
	for i, att := range attempts {
		wipeYtDlpTempArtifacts(tmpDir)
		args := buildYtDlpArgs(tmpDir, youtubePageURL, att.extractor, att.useHints, cookiesPath)
		cmd := exec.Command("yt-dlp", args...)
		cmd.Dir = tmpDir

		slog.Info("running yt-dlp", "attempt", i+1, "videoID", videoID, "extractorArgs", att.extractor, "useHints", att.useHints)
		output, execErr := cmd.CombinedOutput()
		if execErr != nil {
			lastErr = fmt.Errorf("attempt %d: %w (%s)", i+1, execErr, summarizeCmdOutput(output))
			slog.Warn("yt-dlp attempt failed", "attempt", i+1, "err", execErr, "output_tail", summarizeCmdOutput(output))
			continue
		}

		p, ferr := pickMergedOutput(tmpDir)
		if ferr != nil {
			lastErr = ferr
			slog.Warn("yt-dlp output missing after success exit", "attempt", i+1, "detail", ferr)
			continue
		}

		slog.Info("yt-dlp succeeded", "attempt", i+1, "file", p)
		return filepath.Clean(p), cleanupFn, nil
	}

	cleanupFn()
	return "", nil, fmt.Errorf("yt-dlp: barcha urinishlar muvaffaqiyatsiz: %w", lastErr)
}

func extractVideoID(raw string) (string, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", fmt.Errorf("empty youtube url")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	host := strings.ToLower(u.Host)
	host = strings.TrimPrefix(host, "www.")
	var cleanID string

	switch {
	case strings.Contains(host, "youtu.be"):
		cleanID = strings.Trim(strings.TrimPrefix(u.Path, "/"), "/")
	case strings.Contains(host, "youtube.com"), strings.Contains(host, "youtube-nocookie.com"):
		cleanID = u.Query().Get("v")
		if cleanID == "" {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			for i := 0; i < len(parts)-1; i++ {
				switch parts[i] {
				case "embed", "live", "shorts", "v":
					cleanID = parts[i+1]
					i = len(parts)
				default:
				}
			}
		}
	default:
		return "", "", fmt.Errorf("not a supported YouTube URL: %s", raw)
	}

	// parametrlardan qolgan chirchiq
	if idx := strings.IndexAny(cleanID, "?/&"); idx >= 0 {
		cleanID = cleanID[:idx]
	}
	cleanID = strings.TrimSpace(cleanID)
	if cleanID == "" {
		return "", "", fmt.Errorf("could not parse YouTube video id from URL: %s", raw)
	}

	fullID := cleanID
	if u.RawQuery != "" && strings.Contains(host, "youtu.be") {
		fullID = cleanID + "?" + u.RawQuery
	} else if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtube-nocookie.com") {
		// youtube.com uchun 'v' dan tashqari boshqa hamma parametrlarni saqlash mumkin
		q := u.Query()
		q.Del("v")
		if len(q) > 0 {
			fullID = cleanID + "?" + q.Encode()
		}
	}

	return cleanID, fullID, nil
}

type userAgentRoundTripper struct {
	rt http.RoundTripper
}

func (ua userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.google.com")
	resp, err := ua.rt.RoundTrip(req)
	if err == nil && resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("YouTube API Error Details", "status", resp.StatusCode, "body", string(body))
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return resp, err
}
