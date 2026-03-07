package handler

import (
	"context"
	"fmt"
	m "imed/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// SendConfirmation godoc
// @Summary Send confirmation code to email
// @Description Send a confirmation code to the provided email address
// @Tags auth
// @Accept  json
// @Produce  json
// @Param banner body models.SendConfirmationCodeReq true "Code Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/auth/user/otp/send [post]
func (h *Handler) SendConfirmation(c *gin.Context) {
	var req m.SendConfirmationCodeReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	if req.Type != "email" && req.Type != "telegram" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid type, must be 'email' or 'telegram'"})
		slog.Error("Invalid type, must be 'email' or 'telegram'")
		smssender(c, fmt.Errorf("invalid type, must be 'email' or 'telegram'"), http.StatusBadRequest)
		return
	}

	check := isValidIdentifier(req.Identifier)
	if !check {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid identifier format"})
		slog.Error("Invalid identifier format")
		smssender(c, fmt.Errorf("invalid identifier format"), http.StatusBadRequest)
		return
	}

	slog.Info("Identifier:", req)
	res, err := h.Clients.User.SendConfirmationCode(context.Background(), &req)
	slog.Info("Send Confirmation res:", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error while sending otp", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "confirmation code sent successfully"})
}

// VerifyConfirmation godoc
// @Summary Verify confirmation code
// @Description Verify the provided confirmation code and return a token if valid
// @Tags auth
// @Accept  json
// @Produce  json
// @Param banner body models.ConfirmationReq true "Code Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/auth/user/otp/confirm [post]
func (h *Handler) VerifyConfirmation(c *gin.Context) {
	var req m.ConfirmationReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	if req.Identifier != "" && req.Identifier != "string" {

		check := isValidIdentifier(req.Identifier)

		if !check {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid identifier format"})
			slog.Error("Invalid identifier format")
			smssender(c, fmt.Errorf("invalid identifier format"), http.StatusBadRequest)
			return
		}
	}

	if req.ConfirmationCode == "" || req.ConfirmationCode == "string" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "confirmation code is required"})
		slog.Error("Confirmation code is required")
		smssender(c, fmt.Errorf("confirmation code is required"), http.StatusBadRequest)
		return
	}

	if req.Type != "email" && req.Type != "telegram" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid type, must be 'email' or 'telegram'"})
		slog.Error("Invalid type, must be 'email' or 'telegram'")
		smssender(c, fmt.Errorf("invalid type, must be 'email' or 'telegram'"), http.StatusBadRequest)
		return
	}

	slog.Info("Verify Confirmation Code:", req)
	res, err := h.Clients.User.Confirmation(context.Background(), &req)
	slog.Info("Verify Confirmation res:", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error while verifying confirmation code", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}
