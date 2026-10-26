package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const lastSeenThrottle = 5 * time.Minute

// LastSeen mobil foydalanuvchining oxirgi faolligini yozadi (T-04/T-05).
// DB ga yozish Redis orqali har foydalanuvchi uchun 5 daqiqada bir martaga cheklanadi.
// NewAuth dan keyin ulanadi.
func LastSeen(rdb *redis.Client, touch func(ctx context.Context, userID string) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v, _ := ctx.Get("user_id")
		userID, _ := v.(string)
		if userID != "" {
			ok, err := rdb.SetNX(ctx.Request.Context(), "last_seen:"+userID, 1, lastSeenThrottle).Result()
			if err != nil {
				slog.Error("last seen: redis failed", "err", err)
			} else if ok {
				if err := touch(ctx.Request.Context(), userID); err != nil {
					slog.Error("last seen: update failed", "err", err, "user_id", userID)
				}
			}
		}
		ctx.Next()
	}
}
