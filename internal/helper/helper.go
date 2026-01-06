package helper

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"imed/internal/config"
)

func SendMessage(message string) error {
	cf := config.Load()
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cf.BOT_TOKEN)
	params := url.Values{}
	params.Add("chat_id", cf.GROUP_ID)
	params.Add("text", message)

	// So'rov yuborish
	resp, err := http.PostForm(apiURL, params)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Javobni o'qish
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Status kodni tekshirish
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GeneratePassword() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
