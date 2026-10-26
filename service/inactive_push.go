package service

import (
	"context"
	"log/slog"
	"time"
)

const (
	inactiveAfter     = 24 * time.Hour
	inactiveBatchSize = 500
)

var inactivePushText = map[string][2]string{
	"uz": {"TibMentor", "Bugun hali dars qilmadingiz. Davom etamizmi?"},
	"ru": {"TibMentor", "Сегодня вы ещё не занимались. Продолжим?"},
	"en": {"TibMentor", "You haven't studied today yet. Shall we continue?"},
}

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
	for {
		users, err := s.storage.UserS.GetInactiveUsers(ctx, time.Now().Add(-inactiveAfter), inactiveBatchSize)
		if err != nil {
			slog.Error("inactive reminder: get users failed", "err", err)
			return
		}
		if len(users) == 0 {
			return
		}

		byLang := map[string][]string{}
		ids := make([]string, 0, len(users))
		for _, u := range users {
			lang := u.Language
			if _, ok := inactivePushText[lang]; !ok {
				lang = "uz"
			}
			byLang[lang] = append(byLang[lang], u.FCMToken)
			ids = append(ids, u.Id)
		}

		for lang, tokens := range byLang {
			text := inactivePushText[lang]
			data := map[string]string{"type": "inactive_reminder"}
			if _, err := s.push.SendToManyTokens(ctx, tokens, text[0], text[1], data); err != nil {
				slog.Error("inactive reminder: send failed", "lang", lang, "err", err)
			}
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
