package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func run(t *testing.T, status int, path string, n int) string {
	t.Helper()
	requestCounter.Store(0)

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logger())
	r.GET(path, func(c *gin.Context) { c.Status(status) })

	for i := 0; i < n; i++ {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	return buf.String()
}

func TestSmoke(t *testing.T) {
	// 200 -> har 10-inchisi loglanadi.
	if got := strings.Count(run(t, http.StatusOK, "/x", 100), "http request"); got != 10 {
		t.Errorf("200 sampling: want 10, got %d", got)
	}
	// 500 -> hammasi loglanadi.
	if got := strings.Count(run(t, http.StatusInternalServerError, "/x", 25), "http request"); got != 25 {
		t.Errorf("500: want 25, got %d", got)
	}
	// 404 darajasi WARN.
	if out := run(t, http.StatusNotFound, "/x", 1); !strings.Contains(out, "level=WARN") {
		t.Errorf("404 want WARN, got %q", out)
	}
	// /health butunlay o'tkazib yuboriladi.
	if out := run(t, http.StatusOK, "/health", 50); out != "" {
		t.Errorf("/health want no logs, got %q", out)
	}
}
