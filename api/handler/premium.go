package handler

import (
	"context"
	"errors"
	"net/http"

	m "imed/pkg/models"
	"imed/storage/postgres"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// GetPremiumSettings godoc
// @Summary Get premium settings
// @Tags premium
// @Produce json
// @Success 200 {object} models.PremiumSettings
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/premium/settings [get]
func (h *Handler) GetPremiumSettings(c *gin.Context) {
	res, err := h.Clients.Premium.GetSettings(context.Background())
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		smssender(c, err, http.StatusBadRequest)
		return
	}
	c.JSON(200, res)
}

// UpdatePremiumSettings godoc
// @Summary Update premium price and duration
// @Description Price 0 disables buying premium
// @Tags premium
// @Accept json
// @Produce json
// @Param settings body models.PremiumSettings true "Premium settings"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/premium/settings [put]
func (h *Handler) UpdatePremiumSettings(c *gin.Context) {
	body := m.PremiumSettings{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if _, err := h.Clients.Premium.UpdateSettings(context.Background(), &body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "premium settings updated successfully"})
}

// SetUserUstoz godoc
// @Summary Grant or revoke "Ustoz" title
// @Description Ustoz gets permanent access to all courses
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param body body models.UstozBody true "Ustoz flag"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /web/user/{id}/ustoz [put]
func (h *Handler) SetUserUstoz(c *gin.Context) {
	body := m.UstozBody{}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		return
	}
	if _, err := h.Clients.Premium.SetUstoz(context.Background(), &m.UstozReq{UserId: c.Param("id"), IsUstoz: body.IsUstoz}); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		smssender(c, err, http.StatusBadRequest)
		return
	}
	c.JSON(200, gin.H{"message": "ustoz updated successfully"})
}

// GetPremiumInfo godoc
// @Summary Premium price and current user status
// @Tags premium
// @Produce json
// @Success 200 {object} models.PremiumInfoRes
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /mobile/premium [get]
func (h *Handler) GetPremiumInfo(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	res, err := h.Clients.Premium.GetInfo(context.Background(), userId)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		smssender(c, err, http.StatusBadRequest)
		return
	}
	c.JSON(200, res)
}

// CreatePremiumOrder godoc
// @Summary Buy monthly premium via Click
// @Description Paying again extends the active subscription
// @Tags premium
// @Accept json
// @Produce json
// @Param body body models.PremiumOrderBody false "Optional promocode"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Security BearerAuth
// @Router /mobile/premium/order [post]
func (h *Handler) CreatePremiumOrder(c *gin.Context) {
	userId := claimData(c, "user_id")
	if userId == "" {
		return
	}
	body := m.PremiumOrderBody{}
	if c.Request.ContentLength > 0 {
		if err := c.BindJSON(&body); err != nil {
			c.JSON(400, gin.H{"message": err.Error()})
			return
		}
	}

	res, err := h.Clients.Premium.CreateOrder(context.Background(), &m.PremiumOrderReq{UserId: userId, PromoCodeId: body.PromoCodeId})
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		if !errors.Is(err, postgres.ErrPremiumNotConfigured) {
			slog.Error("Error creating premium order: ", err)
			smssender(c, err, http.StatusBadRequest)
		}
		return
	}

	c.JSON(200, gin.H{
		"message":  "order created successfully",
		"order_id": res.Id,
		"url":      res.Url,
	})
}
