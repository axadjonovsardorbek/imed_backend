package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"imed/api/token"

	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt"

	"github.com/gin-gonic/gin"
)

const (
	key          = "sardorbek"
	unauthorized = "unauthorized"
)

func NewAuth(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		allow, err := CheckPermission(ctx.FullPath(), ctx.Request, enforcer)

		if err != nil {
			slog.Error("Error checking permission: %v", "err", err)
			if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
				RequireRefresh(ctx)
				return
			} else {
				RequirePermission(ctx)
				return
			}
		} else if !allow {
			RequirePermission(ctx)
			return
		}

		claims, err := ExtractToken(ctx.Request)
		if err != nil {
			slog.Error("Error extracting token: %v", "err", err)
			InvalidToken(ctx)
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("user_id", claims["user_id"])

		ctx.Next()
	}
}

func ExtractToken(r *http.Request) (jwt.MapClaims, error) {
	jwtToken := r.Header.Get("Authorization")
	if jwtToken == "" {
		return nil, fmt.Errorf("access token missing")
	}
	if strings.Contains(jwtToken, "Basic") {
		return nil, fmt.Errorf("invalid token format")
	}

	tokenString := strings.TrimPrefix(jwtToken, "Bearer ")
	claims, err := token.ExtractClaim(tokenString)
	if err != nil {
		slog.Error("Error while extracting claims", "err", err)
		return nil, err
	}

	return claims, nil
}

func GetRole(r *http.Request) (string, error) {
	claims, err := ExtractToken(r)
	if err != nil {
		return unauthorized, err
	}

	role, ok := claims["role"].(string)
	if !ok {
		return unauthorized, errors.New("role claim not found")
	}
	return role, nil
}

func CheckPermission(path string, r *http.Request, enforcer *casbin.Enforcer) (bool, error) {
	role, err := GetRole(r)
	if err != nil {
		slog.Error("Error getting role from token", "err", err)
		return false, err
	}

	allowed, err := enforcer.Enforce(role, path, r.Method)
	if err != nil {
		slog.Error("Error during Casbin enforce", "err", err)
		return false, err
	}

	return allowed, nil
}

func InvalidToken(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": "Invalid token!",
	})
}

func RequirePermission(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": "Permission denied",
	})
}

func RequireRefresh(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": "Access token expired",
	})
}
