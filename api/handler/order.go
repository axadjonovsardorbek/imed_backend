package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	m "imed/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// Create godoc
// @Summary Create a new order
// @Description Create a new order with the provided details
// @Tags order
// @Accept  json
// @Produce  json
// @Param order body models.OrderCreateBody true "Order Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/order [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	req := m.OrderCreateReq{}
	body := m.OrderCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req.UserId = user_id
	req.CourseId = body.CourseId
	req.TariffId = body.TariffId

	slog.Info("Create order", "order", req)
	res, err := h.Clients.Order.Create(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error creating order: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"message":  "order created successfully",
		"order_id": res.Id,
		"url":      res.Url,
	})
}

func (h *Handler) GetPayURL(c *gin.Context) {
	orderID := c.Param("id")

	ord, err := h.Clients.Order.GetById(context.Background(), &m.ById{Id: orderID})
	if err != nil {
		c.JSON(404, gin.H{"message": err.Error()})
		return
	}
	if ord.Status == "PAID" {
		c.JSON(409, gin.H{"message": "already paid"})
		return
	}

	payURL := fmt.Sprintf(
		"https://my.click.uz/services/pay?merchant_id=%d&service_id=%d&amount=%.2f&transaction_param=%s&return_url=%s",
		h.Cfg.CLICK_MERCHANT_ID,
		h.Cfg.CLICK_SERVICE_ID,
		ord.Amount,
		orderID,
		url.QueryEscape(h.Cfg.BASE_URL+"/pay/return"),
	)

	c.JSON(200, gin.H{"order_id": orderID, "pay_url": payURL})
}
