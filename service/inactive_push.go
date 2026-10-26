package service

import (
	"context"
	"log/slog"
	"time"

	m "imed/pkg/models"
)

const (
	inactiveAfter     = 24 * time.Hour
	inactiveBatchSize = 500
)

// RunInactiveReminder 24 soat kirmagan foydalanuvchilarga eslatma yuboradi (T-05).
// Har bir inaktivlik davrida bitta push ketadi: foydalanuvchi qaytib kirsa
// last_seen_at yangilanadi va hisob qaytadan boshlanadi.
func (s *PushService) RunInactiveReminder(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		s.sendInactiveReminders(ctx)

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *PushService) sendInactiveReminders(ctx context.Context) {
	event := m.PushEvent{Type: m.EventInactiveReminder}

	for {
		users, err := s.storage.UserS.GetInactiveUsers(ctx, time.Now().Add(-inactiveAfter), inactiveBatchSize)
		if err != nil {
			slog.Error("inactive reminder: get users failed", "err", err)
			return
		}
		if len(users) == 0 {
			return
		}

		ids := make([]string, 0, len(users))
		for _, u := range users {
			// To'lov pushlari kabi user_id orqali yuboriladi.
			if err := s.SendPush(ctx, m.PushTarget{Type: m.TargetUser, UserID: u.Id}, event); err != nil {
				slog.Error("inactive reminder: send failed", "user_id", u.Id, "err", err)
			}
			ids = append(ids, u.Id)
		}

		// Yuborish xato bo'lsa ham belgilanadi, aks holda har tickda qayta urinib spam bo'ladi.
		if err := s.storage.UserS.MarkInactivePushed(ctx, ids); err != nil {
			slog.Error("inactive reminder: mark failed", "err", err)
			return
		}
		slog.Info("inactive reminder sent", "count", len(ids))

		if len(users) < inactiveBatchSize {
			return
		}
	}
}
