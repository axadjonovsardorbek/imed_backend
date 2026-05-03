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
	filePath, err := downloadYouTube(youtubeURL, cleanID)
	if err != nil {
		slog.Error("Failed to download video from YouTube", "err", err)
		return "", "", err
	}
	slog.Info("Video downloaded successfully", "path", filePath)
	defer os.Remove(filePath)

	// 3. upload to S3 (Contabo)
	url, err := s.UploadToS3(filePath, cleanID)
	if err != nil {
		slog.Error("Failed to upload video to S3", "err", err, "path", filePath)
		return "", "", err
	}
	slog.Info("Video uploaded to S3 successfully", "url", url)

	return url, fullID, nil
}

func downloadYouTube(videoURL, videoID string) (string, error) {
	fileName := "tmp_" + videoID + ".mp4"
	slog.Info("Downloading video using yt-dlp", "videoID", videoID, "output", fileName)

	// yt-dlp buyrug'ini shakllantiramiz
	args := []string{
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"-o", fileName,
		"--extractor-args", "youtube:player_client=android",
		"--no-check-certificate",
	}

	// Agar cookies.txt fayli mavjud bo'lsa, undan foydalanamiz
	if _, err := os.Stat("cookies.txt"); err == nil {
		slog.Info("Using cookies.txt for yt-dlp")
		args = append(args, "--cookies", "cookies.txt")
	}

	args = append(args, "https://www.youtube.com/watch?v="+videoID)

	cmd := exec.Command("yt-dlp", args...)

	// Xatoliklarni ko'rish uchun outputni bog'laymiz
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("yt-dlp failed", "err", err, "output", string(output))
		return "", fmt.Errorf("yt-dlp error: %v, output: %s", err, string(output))
	}

	slog.Info("yt-dlp finished successfully")
	return fileName, nil
}

func extractVideoID(raw string) (string, string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}

	var cleanID, fullID string

	if strings.Contains(u.Host, "youtu.be") {
		cleanID = strings.TrimPrefix(u.Path, "/")
	} else {
		cleanID = u.Query().Get("v")
	}

	fullID = cleanID
	if u.RawQuery != "" && !strings.Contains(u.Host, "youtube.com") {
		// youtu.be uchun query parametrlarni saqlab qolamiz
		fullID = cleanID + "?" + u.RawQuery
	} else if strings.Contains(u.Host, "youtube.com") {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://www.youtube.com")
	resp, err := ua.rt.RoundTrip(req)
	if err == nil && resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("YouTube API Error Details", "status", resp.StatusCode, "body", string(body))
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return resp, err
}
