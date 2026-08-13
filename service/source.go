package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"imed/internal/helper"
	m "imed/pkg/models"
	"imed/pkg/s3"
	st "imed/storage/postgres"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"log/slog"
)

type SourceService struct {
	storage st.Storage
	rdb     *redis.Client
	push    *PushService
	s3      *s3.S3Service
}

func NewSourceService(storage *st.Storage, rdb *redis.Client, push *PushService, s3 *s3.S3Service) *SourceService {
	return &SourceService{storage: *storage, rdb: rdb, push: push, s3: s3}
}

func (s *SourceService) getVersion() string {
	version := "1"
	if s.rdb != nil {
		if v, err := s.rdb.Get("source:version").Result(); err == nil && v != "" {
			version = v
		}
	}
	return version
}

func (s *SourceService) Create(ctx context.Context, req *m.SourceCreateReq) (*m.ById, error) {
	res, err := s.storage.SourceS.Create(req)
	if err != nil {
		return nil, err
	}

	// 🔥 cache invalidation
	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	// GetLessonCourseInfo lessons.id kutadi — source id emas!
	info, infoErr := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: req.LessonId})
	if infoErr != nil {
		slog.Warn("source Create: skipped push/notifications — GetLessonCourseInfo failed", "lesson_id", req.LessonId, "source_id", res.Id, "err", infoErr)
	} else if s.push != nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}
		push_event := m.PushEvent{
			Type:       m.EventSourceAdded,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
			SourceType: req.Type,
		}
		if err := s.push.SendPush(ctx, push_target, push_event); err != nil {
			slog.Error("source Create: SendPush failed", "err", err, "lesson_id", req.LessonId)
		}
		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)
		if _, nerr := s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		}); nerr != nil {
			slog.Error("source Create: notification write failed", "err", nerr)
		}
	}

	if req.Type == "video" && s.s3 != nil {
		s.maybeProcessYouTubeVideo(res.Id, req.Url, req.VideoUrl)
	}

	return res, nil
}

func (s *SourceService) maybeProcessYouTubeVideo(sourceID, urlJSON, videoURLJSON string) {
	if findYouTubeURL(urlJSON, videoURLJSON) == "" {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("ProcessYouTubeVideo panic", "sourceID", sourceID, "recover", r)
			}
		}()
		s.ProcessYouTubeVideo(sourceID, urlJSON, videoURLJSON)
	}()
}

func (s *SourceService) GetById(ctx context.Context, req *m.ById) (*m.SourceRes, error) {
	return s.storage.SourceS.GetById(req)
}

func (s *SourceService) GetAll(ctx context.Context, req *m.SourceGetListReq) (*m.PaginationResponse, *m.SourceList, error) {

	version := s.getVersion()

	key := fmt.Sprintf(
		"source:list:v%s_lesson:%s_name:%s_type:%s_page:%d_limit:%d",
		version,
		req.LessonId,
		req.Name,
		req.Type,
		req.Filter.Page,
		req.Filter.Limit,
	)

	if s.rdb != nil {
		if b, err := s.rdb.Get(key).Bytes(); err == nil && len(b) > 0 {
			var cached m.PaginationResponse
			if err := json.Unmarshal(b, &cached); err == nil {
				return &cached, nil, nil
			}
		}
	}

	res, list, err := s.storage.SourceS.GetAll(req)
	if err != nil {
		return nil, nil, err
	}

	if s.rdb != nil {
		if b, err := json.Marshal(res); err == nil {
			_ = s.rdb.Set(key, b, 5*time.Minute).Err()
		}
	}

	return res, list, nil
}

func (s *SourceService) Update(ctx context.Context, req *m.SourceUpdateReq) (*m.Void, error) {
	var prevYouTube string
	if existing, err := s.storage.SourceS.GetById(&m.ById{Id: req.Id}); err == nil {
		prevURLJSON, _ := json.Marshal(existing.Url)
		prevVideoURLJSON, _ := json.Marshal(existing.VideoUrl)
		prevYouTube = findYouTubeURL(string(prevURLJSON), string(prevVideoURLJSON))
	}

	res, err := s.storage.SourceS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	info, infoErr := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: req.LessonId})
	if infoErr != nil {
		slog.Warn("source Update: skipped push/notifications — GetLessonCourseInfo failed", "lesson_id", req.LessonId, "source_id", req.Id, "err", infoErr)
	} else if s.push != nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}
		push_event := m.PushEvent{
			Type:       m.EventSourceUpdated,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
			SourceType: req.Type,
		}
		if err := s.push.SendPush(ctx, push_target, push_event); err != nil {
			slog.Error("source Update: SendPush failed", "err", err, "lesson_id", req.LessonId)
		}
		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)
		if _, nerr := s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		}); nerr != nil {
			slog.Error("source Update: notification write failed", "err", nerr)
		}
	}

	if req.Type == "video" && s.s3 != nil {
		newYouTube := findYouTubeURL(req.Url, req.VideoUrl)
		if newYouTube != "" && newYouTube != prevYouTube {
			s.maybeProcessYouTubeVideo(req.Id, req.Url, req.VideoUrl)
		}
	}

	return res, nil
}


func (s *SourceService) Delete(ctx context.Context, req *m.ById) (*m.Void, error) {
	res, err := s.storage.SourceS.Delete(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	return res, nil
}

func parseSourceURLMap(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var urls map[string]string
	if err := json.Unmarshal([]byte(raw), &urls); err != nil {
		if strings.HasPrefix(strings.ToLower(raw), "http") {
			return map[string]string{"direct": strings.Trim(raw, "\"")}
		}
		return nil
	}
	return urls
}

func findYouTubeURL(urlSources ...string) string {
	for _, src := range urlSources {
		for _, u := range parseSourceURLMap(src) {
			u = strings.TrimSpace(u)
			if u != "" && (strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be")) {
				return u
			}
		}
	}
	return ""
}

// truncateForTelegram — xato matnini Telegram xabariga sig'adigan qilib qisqartiradi
// va HTML parse mode buzilmasligi uchun maxsus belgilarni ekranlaydi.
// yt-dlp chiqishi ancha uzun bo'lishi mumkin; eng muhim qismi boshida (sabab tegi) turadi.
func truncateForTelegram(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max] + " …"
	}
	return html.EscapeString(s)
}

func (s *SourceService) ProcessYouTubeVideo(sourceID, urlJSON, videoURLJSON string) {
	youtubeURL := findYouTubeURL(urlJSON, videoURLJSON)
	if youtubeURL == "" {
		slog.Warn("no youtube url in source url/video_url fields", "sourceID", sourceID)
		return
	}

	// Telegramga xabar yuboramiz
	logMsg := fmt.Sprintf("🎥 <b>YouTube yuklash boshlandi</b>\n\n🔗 <b>Link:</b> %s\n🆔 <b>Source ID:</b> %s", youtubeURL, sourceID)
	_ = helper.SendMessage(logMsg)

	slog.Info("Starting background YouTube processing", "sourceID", sourceID, "url", youtubeURL)

	videoURL, _, err := s.s3.UploadFromYouTube(youtubeURL)
	if err != nil {
		slog.Error("Failed to process YouTube video in background", "err", err, "sourceID", sourceID)
		errorMsg := fmt.Sprintf("❌ <b>YouTube yuklashda xatolik!</b>\n\n🆔 <b>Source ID:</b> %s\n🔗 <b>Link:</b> %s\n⚠️ <b>Xato:</b> %v", sourceID, youtubeURL, truncateForTelegram(err.Error(), 900))
		_ = helper.SendMessage(errorMsg)
		return
	}

	// Bazani yangilaymiz. Hamma tillar uchun bir xil S3 linkini qo'yamiz.
	videoUrls := map[string]string{
		"uz": videoURL,
		"ru": videoURL,
		"en": videoURL,
	}

	videoUrlsJSON, _ := json.Marshal(videoUrls)
	err = s.storage.SourceS.UpdateVideoUrl(sourceID, string(videoUrlsJSON))
	if err != nil {
		slog.Error("Failed to update video_url in database", "err", err, "sourceID", sourceID)
		return
	}

	slog.Info("Successfully processed and updated YouTube video", "sourceID", sourceID, "videoURL", videoURL)

	successMsg := fmt.Sprintf("✅ <b>Video muvaffaqiyatli yuklandi!</b>\n\n🆔 <b>Source ID:</b> %s\n📥 <b>S3 Link:</b> <a href=\"%s\">Video Link</a>", sourceID, videoURL)
	_ = helper.SendMessage(successMsg)

	// Cache invalidation again after update
	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}
}
