package handler

import (
	"context"
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
// @Router /web/auth/user/otp/send [post]
func (h *Handler) SendConfirmation(c *gin.Context) {
	var req m.SendConfirmationCodeReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Email:", req)
	res, err := h.Clients.User.SendConfirmationCode(context.Background(), &req)
	slog.Info("Send Confirmation res:", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error while login", err)
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
// @Router /web/auth/user/otp/confirm [post]
func (h *Handler) VerifyConfirmation(c *gin.Context) {
	var req m.ConfirmationReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
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
