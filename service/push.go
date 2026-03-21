package service

import (
	"context"
	"fmt"
	"imed/internal/push"
	m "imed/pkg/models"
	st "imed/storage/postgres"
)

type PushService struct {
	storage st.Storage
	push    *push.Client
}

func NewPushService(storage *st.Storage, push *push.Client) *PushService {
	return &PushService{storage: *storage, push: push}
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
		return s.storage.UserS.GetTokens(ctx, m.TokenFilter{})

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
		if *token == "" {
			return []string{}, nil
		}
		return []string{*token}, nil

	default:
		return nil, fmt.Errorf("unknown target")
	}
}

func (s *PushService) SendPush(ctx context.Context, target m.PushTarget, event m.PushEvent) error {
	tokens, err := s.ResolveTokens(ctx, target)
	if err != nil {
		fmt.Println("Failed to resolve tokens:", err)
		return err
	}
	if len(tokens) == 0 {
		fmt.Println("No tokens found for target:", target)
		return nil
	}

	msg := m.BuildUzPushMessage(event)

	for _, part := range chunkStrings(tokens, 500) {
		if _, err := s.push.SendToManyTokens(ctx, part, msg.Title, msg.Body, msg.Data); err != nil {
			fmt.Println("Failed to send push to tokens:", err)
			return err
		}
	}
	return nil
}
