package service

import (
	"context"
	"fmt"
	"imed/internal/helper"
	"imed/internal/push"
	m "imed/pkg/models"
	st "imed/storage/postgres"
	"log/slog"
)

type PushService struct {
	storage         st.Storage
	push            *push.Client
	autoContentPush bool
}

func NewPushService(storage *st.Storage, push *push.Client, autoContentPush bool) *PushService {
	return &PushService{storage: *storage, push: push, autoContentPush: autoContentPush}
}

// contentEvents kontent qo'shilganda/o'zgarganda avtomatik ketadigan pushlar (T-17).
var contentEvents = map[m.PushEventType]bool{
	m.EventNewCourse:      true,
	m.EventLessonAdded:    true,
	m.EventModuleAdded:    true,
	m.EventSourceAdded:    true,
	m.EventCourseUpdated:  true,
	m.EventModuleUpdated:  true,
	m.EventLessonUpdated:  true,
	m.EventSourceUpdated:  true,
	m.EventBannerAdded:    true,
	m.EventSubjectAdded:   true,
	m.EventSubjectUpdated: true,
}

func chunkStrings(in []string, size int) [][]string {
	if size <= 0 {
		size = 500
	}
	out := make([][]string, 0, (len(in)+size-1)/size)
	for i := 0; i < len(in); i += size {
		j := i + size
		if j > len(in) {
			j = len(in)
		}
		out = append(out, in[i:j])
	}
	return out
}

func (s *PushService) ResolveTokens(ctx context.Context, t m.PushTarget) ([]string, error) {
	switch t.Type {
	case m.TargetAllUsers:
		// Endilikda TargetAllUsers topic orqali ishlaydi, shuning uchun token qaytarmaymiz
		return []string{}, nil

	case m.TargetTopic:
		// Topic uchun token kerak emas, to‘g‘ridan-to‘g‘ri topic’ga yuboramiz
		return []string{}, nil

	case m.TargetCourseUsers:
		if t.CourseID == "" {
			return nil, fmt.Errorf("course_id required")
		}
		return s.storage.UserS.GetTokens(ctx, m.TokenFilter{CourseID: t.CourseID})

	case m.TargetUser:
		if t.UserID == "" {
			return nil, fmt.Errorf("user_id required")
		}
		token, err := s.storage.UserS.GetToken(t.UserID)
		if err != nil {
			return nil, err
		}
		if token == nil || *token == "" {
			return []string{}, nil
		}
		return []string{*token}, nil

	default:
		return nil, fmt.Errorf("unknown target")
	}
}

func (s *PushService) SendPush(ctx context.Context, target m.PushTarget, event m.PushEvent) error {
	if !s.autoContentPush && contentEvents[event.Type] {
		slog.Debug("auto content push disabled, skipped", "event", event.Type)
		return nil
	}
	if target.Type == m.TargetTopic || target.Type == m.TargetAllUsers {
		topics := []string{}
		if target.Type == m.TargetTopic && target.Topic != "" {
			topics = append(topics, target.Topic)
		} else {
			topics = []string{"news_uz", "news_ru", "news_en"}
		}

		for _, topic := range topics {
			var msg m.PushMessage
			lang := "uz"
			if topic == "news_ru" {
				lang = "ru"
			} else if topic == "news_en" {
				lang = "en"
			}

			if (event.Type == m.EventAdminNotification || event.Type == m.EventBannerAdded || event.Type == m.EventSubjectAdded || event.Type == m.EventSubjectUpdated) && event.TitleJSON != "" {
				event.Title, _ = helper.PickFromJSONStr(event.TitleJSON, lang)
				event.Body, _ = helper.PickFromJSONStr(event.MessageJSON, lang)
			}

			if lang == "ru" {
				msg = m.BuildRuPushMessage(event)
			} else if lang == "en" {
				msg = m.BuildEnPushMessage(event)
			} else {
				msg = m.BuildUzPushMessage(event)
			}

			err := s.push.SendToTopic(ctx, topic, msg.Title, msg.Body, msg.Data)
			if err != nil {
				slog.Error("push to topic failed", "topic", topic, "err", err)
			}
		}
		return nil
	}

	tokens, err := s.ResolveTokens(ctx, target)
	if err != nil {
		slog.Error("resolve push tokens failed", "err", err)
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	if (event.Type == m.EventAdminNotification || event.Type == m.EventBannerAdded || event.Type == m.EventSubjectAdded || event.Type == m.EventSubjectUpdated) && event.TitleJSON != "" {
		event.Title, _ = helper.PickFromJSONStr(event.TitleJSON, "uz")
		event.Body, _ = helper.PickFromJSONStr(event.MessageJSON, "uz")
	}

	msg := m.BuildUzPushMessage(event)

	for _, part := range chunkStrings(tokens, 500) {
		res, err := s.push.SendToManyTokens(ctx, part, msg.Title, msg.Body, msg.Data)
		if len(res.InvalidTokens) > 0 {
			_, _ = s.storage.UserS.DeleteToken(res.InvalidTokens)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
