package token

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
)

const (
	signingKey = "sardorbek"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateJWTToken(userID string, role string, rdb *redis.Client) (*Tokens, error) {
	accessToken := jwt.New(jwt.SigningMethodHS256)
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	claims := accessToken.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["role"] = role
	claims["session_id"] = sessionID
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(4320 * time.Hour).Unix() // Token expires in 180 days
	access, err := accessToken.SignedString([]byte(signingKey))
	if err != nil {
		return nil, fmt.Errorf("error while generating access token: %w", err)
	}

	rftClaims := refreshToken.Claims.(jwt.MapClaims)
	rftClaims["user_id"] = userID
	rftClaims["role"] = role
	rftClaims["session_id"] = sessionID
	rftClaims["iat"] = time.Now().Unix()
	rftClaims["exp"] = time.Now().Add(4320 * time.Hour).Unix() // Refresh token expires in 180 days
	refresh, err := refreshToken.SignedString([]byte(signingKey))
	if err != nil {
		return nil, fmt.Errorf("error while generating refresh token: %w", err)
	}

	redisKey := "user_session:" + userID
	// Session yozuvi login uchun majburiy emas: xato bo'lsa faqat log qilinadi,
	// token baribir qaytariladi (avvalgi xatti-harakat saqlanadi).
	if err := rdb.Set(context.Background(), redisKey, sessionID, 4320*time.Hour).Err(); err != nil { // Session expires in 180 days
		slog.Error("error while saving user session", "user_id", userID, "err", err)
	}

	return &Tokens{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func ValidateToken(tokenStr string) (bool, error) {
	_, err := ExtractClaim(tokenStr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func ExtractClaim(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
