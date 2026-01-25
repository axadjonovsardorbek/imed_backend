package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	sms "imed/internal/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

const phoneRegex = `^(\+998)?[0-9]{9}$`
const emailRegex = `^[a-zA-Z0-9._]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

func isValidEmail(email string) bool {
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

func claimData(c *gin.Context, data string) string {
	claims, exists := c.Get("claims")
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

func IsYYYYMMDD(s string) bool {
	if len(s) != 10 {
		return false
	}
	// tezkor tekshiruv: YYYY-MM-DD
	if s[4] != '-' || s[7] != '-' {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
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
	fileName := id + header.Filename

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
