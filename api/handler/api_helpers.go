package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	sms "imed/internal/helper"
	"imed/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

const semverRegex = `^\d+\.\d+\.\d+$`
const phoneRegex = `^\+?[1-9]\d{7,14}$`
const emailRegex = `^[a-zA-Z0-9._]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

func isValidEmail(email string) bool {
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidIdentifier(input string) bool {
	return isValidEmail(input) || isValidPhone(input)
}

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

func isValidSemver(v string) bool {
	re := regexp.MustCompile(semverRegex)
	return re.MatchString(v)
}

func isValidAppVersion(v models.AppVersionRes) bool {
	return isValidSemver(v.Android) && isValidSemver(v.IOS)
}

func isValidAppLinks(l models.AppLinksRes) bool {
	return isValidURL(l.Google) && isValidURL(l.Apple) && (l.ApkUrl == "" || isValidURL(l.ApkUrl))
}

func isValidURL(u string) bool {
	parsed, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func validateAppRouteCreate(req models.AppRouteCreateReq) error {

	if !isValidPhone(req.CallCenter) {
		return fmt.Errorf("invalid call_center phone number")
	}

	if !isValidURL(req.SupportUrl) {
		return fmt.Errorf("invalid support_url")
	}

	if !isValidURL(req.FeedbackUrl) {
		return fmt.Errorf("invalid feedback_url")
	}

	if !isValidAppVersion(req.AppVersion) {
		return fmt.Errorf("invalid app_version format")
	}

	if !isValidAppLinks(req.AppLinks) {
		return fmt.Errorf("invalid app_links")
	}

	if !isValidSemver(req.PaymentMinVersion) {
		return fmt.Errorf("invalid payment_min_version")
	}

	return nil
}

func claimData(c *gin.Context, data string) string {
	claims, exists := c.Get("claims")
	fmt.Println("claims", claims)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		smssender(c, errors.New("error: Unauthorized"), 401)
		return ""
	}

	res := claims.(jwt.MapClaims)[data].(string)

	return res
}

func smssender(c *gin.Context, err error, code int) {
	url := c.Request.URL
	message := fmt.Sprint("Status code: ", code, "\nEndpoint: ", url, "\nError: ", err)
	err = sms.SendMessage(message)
	if err == nil {
		slog.Info("Error is successfully sent to group by bot")
	} else {
		slog.Error("Error isnt sending to group by bot: ", err)
	}
}

// func IsYYYYMMDD(s string) bool {
// 	if len(s) != 10 {
// 		return false
// 	}
// 	// tezkor tekshiruv: YYYY-MM-DD
// 	if s[4] != '-' || s[7] != '-' {
// 		return false
// 	}
// 	_, err := time.Parse("2006-01-02", s)
// 	return err == nil
// }

func IsYYYYMMDD(s string) bool {
	switch len(s) {
	case 10:
		_, err := time.Parse("2006-01-02", s)
		return err == nil
	default:
		_, err := time.Parse(time.RFC3339, s)
		return err == nil
	}
}

func getLanguage(c *gin.Context) string {

	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		lang = "uz"
	}

	switch lang {
	case "uz", "ru", "en":
	default:
		lang = "uz"
	}

	return lang
}

func formStr(c *gin.Context, key string) string {
	return strings.TrimSpace(c.PostForm(key))
}

func formFile(c *gin.Context, key string) (string, string, error) {
	file, header, err := c.Request.FormFile(key)

	if file == nil {
		return "", "", nil
	}

	if err != nil {
		return "", "", fmt.Errorf("file not provided")
	}
	defer file.Close()

	id := uuid.New().String()

	// Fayl nomini olish
	ext := filepath.Ext(header.Filename)

	fileName := id + ext

	// Faylni vaqtinchalik joyga saqlash
	tempFilePath := filepath.Join("/tmp", fileName)
	out, err := os.Create(tempFilePath)
	if err != nil {
		return "", "", fmt.Errorf("unable to create temporary file")
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", "", fmt.Errorf("failed to save file")
	}

	return fileName, tempFilePath, nil
}
