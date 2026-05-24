package service

import (
	"context"
	"encoding/json"
	"fmt"
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

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: res.Id})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		push_event := m.PushEvent{
			Type:       m.EventSourceAdded,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
			SourceType: req.Type,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		})
	}

	if req.Type == "video" && s.s3 != nil {
		go s.ProcessYouTubeVideo(res.Id, req.Url)
	}

	return res, nil
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
	res, err := s.storage.SourceS.Update(req)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		_ = s.rdb.Incr("source:version").Err()
	}

	info, err := s.storage.LessonS.GetLessonCourseInfo(&m.ById{Id: req.Id})
	if err == nil {
		push_target := m.PushTarget{Type: m.TargetAllUsers}

		push_event := m.PushEvent{
			Type:       m.EventSourceUpdated,
			CourseID:   info.CourseId,
			CourseName: info.CourseName,
			LessonID:   info.LessonId,
			LessonName: info.LessonName,
			SourceType: req.Type,
		}

		_ = s.push.SendPush(ctx, push_target, push_event)

		msg := m.BuildUzPushMessage(push_event)
		title, body := helper.StringToJSON(msg.Title, msg.Body)

		_, _ = s.storage.NotificationS.Create(&m.NotificationCreateReq{
			Title:   title,
			Message: body,
			Type:    "all",
		})
	}

	if req.Type == "video" && s.s3 != nil {
		go s.ProcessYouTubeVideo(req.Id, req.Url)
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

func (s *SourceService) ProcessYouTubeVideo(sourceID, urlJSON string) {
	var urls map[string]string
	err := json.Unmarshal([]byte(urlJSON), &urls)
	if err != nil {
		slog.Error("Failed to unmarshal youtube urls", "err", err)
		return
	}

	// Birinchi mavjud linkni olamiz (uz, ru yoki en)
	var youtubeURL string
	for _, u := range urls {
		if u != "" && (strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be")) {
			youtubeURL = u
			break
		}
	}

	if youtubeURL == "" {
		return
	}

	// Telegramga xabar yuboramiz
	logMsg := fmt.Sprintf("🎥 <b>YouTube yuklash boshlandi</b>\n\n🔗 <b>Link:</b> %s\n🆔 <b>Source ID:</b> %s", youtubeURL, sourceID)
	_ = helper.SendMessage(logMsg)

	slog.Info("Starting background YouTube processing", "sourceID", sourceID, "url", youtubeURL)

	videoURL, _, err := s.s3.UploadFromYouTube(youtubeURL)
	if err != nil {
		slog.Error("Failed to process YouTube video in background", "err", err, "sourceID", sourceID)
		errorMsg := fmt.Sprintf("❌ <b>YouTube yuklashda xatolik!</b>\n\n🆔 <b>Source ID:</b> %s\n⚠️ <b>Xato:</b> %v", sourceID, err)
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
