package middleware

import (
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// slowRequest: shundan sekin request har doim loglanadi, sampling'ga tushmaydi.
	slowRequest = time.Second

	// sampleRate: muvaffaqiyatli va tez requestlarning har N-inchisi loglanadi.
	// Trafikning katta qismi shu toifada, log hajmini asosan shu kamaytiradi.
	sampleRate = 10
)

// skipPaths: shovqin yaratadigan, diagnostik qiymati yo'q yo'llar.
var skipPaths = map[string]struct{}{
	"/health": {},
	"/ready":  {},
}

// requestCounter: sampling uchun. rand o'rniga atomik hisoblagich —
// lock yo'q va namuna aniq har N-inchi requestga tushadi.
var requestCounter atomic.Uint64

// Logger: gin.Logger() o'rnini bosadi. Har bir requestni stdout'ga yozish
// o'rniga strukturali slog yozuvi chiqaradi va muvaffaqiyatli requestlarni
// sampling qiladi. Xato va sekin requestlar hech qachon tashlab ketilmaydi.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, skip := skipPaths[path]; skip {
			c.Next()
			return
		}

		start := time.Now()
		raw := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		// Sampling qarori: xato yoki sekin bo'lsa doim, aks holda har N-inchisi.
		if status < 400 && latency < slowRequest {
			if requestCounter.Add(1)%sampleRate != 0 {
				return
			}
		}

		attrs := []any{
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency_ms", latency.Milliseconds(),
		}
		if raw != "" {
			attrs = append(attrs, "query", raw)
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			attrs = append(attrs, "errors", errs)
		}

		switch {
		case status >= 500:
			slog.Error("http request", attrs...)
		case status >= 400:
			slog.Warn("http request", attrs...)
		default:
			slog.Info("http request", attrs...)
		}
	}
}
