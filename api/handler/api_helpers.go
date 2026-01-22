package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	sms "imed/internal/helper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
