package handler

import (
	"context"
	"fmt"
	"net/http"

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
	req.PromoCodeId = body.PromoCodeId

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

// GetOrderById godoc
// @Summary Get order by ID
// @Description Retrieve order details by order ID
// @Tags order
// @Accept  json
// @Produce  json
// @Param id path string true "Order ID"
// @Success 200 {object} models.OrderStatisticRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/order/{id} [get]
func (h *Handler) GetByIdOrder(c *gin.Context) {
	id := c.Param("id")
	req := m.ById{Id: id}

	slog.Info("Get order by ID", "order_id", id)
	res, err := h.Clients.Order.GetById(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting order by ID: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"order": res,
	})
}

// GetAllOrders godoc
// @Summary Get all orders
// @Description Retrieve a list of all orders with optional filters
// @Tags order
// @Accept  json
// @Produce  json
// @Param user_id query string false "User ID"
// @Param course_id query string false "Course ID"
// @Param tariff_id query string false "Tariff ID"
// @Param promocode_id query string false "Promo Code ID"
// @Param status query string false "Order Status" Enums(NEW, PAID, CANCELLED, RESERVED, EXPIRED)
// @Param payment_type query string false "Payment Type" Enums(click, admin)
// @Param type query string false "Type (range, day, week, month, year)" Enums(range, day, week, month, year)
// @Param day query string false "Day (YYYY-MM-DD)"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} models.OrderList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/order [get]
func (h *Handler) GetAllOrders(c *gin.Context) {
	req := m.OrderGetListReq{}

	user_id := c.Query("user_id")
	course_id := c.Query("course_id")
	tariff_id := c.Query("tariff_id")
	promocode_id := c.Query("promocode_id")
	status := c.Query("status")
	payment_type := c.Query("payment_type")
	req_type := c.Query("type")
	day := c.Query("day")
	from := c.Query("from")
	to := c.Query("to")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	if status != "" && status != "string" {
		if status != "NEW" && status != "PAID" && status != "CANCELLED" && status != "RESERVED" && status != "EXPIRED" {
			slog.Error("Invalid status value")
			smssender(c, fmt.Errorf("invalid status value"), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{"message": "status must be one of NEW, PAID, CANCELLED, RESERVED, EXPIRED"})
			return
		}
	}

	if payment_type != "" && payment_type != "string" {
		if payment_type != "click" && payment_type != "admin" {
			slog.Error("Invalid payment_type value")
			smssender(c, fmt.Errorf("invalid payment_type value"), http.StatusBadRequest)
			c.JSON(http.StatusBadRequest, gin.H{"message": "payment_type must be one of click, admin"})
			return
		}
	}

	if req_type != "" {
		if req_type == "day" {
			if day != "" && !IsYYYYMMDD(day) {
				slog.Error("Invalid date format for day type")
				smssender(c, fmt.Errorf("invalid date format for day type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "date must be YYYY-MM-DD"})
				return
			}
			req.Day = day
		}

		if req_type == "range" {
			if from == "" || to == "" {
				slog.Error("from and to are required for range type")
				smssender(c, fmt.Errorf("from and to are required for range type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "from and to are required for range"})
				return
			}
			if !IsYYYYMMDD(from) || !IsYYYYMMDD(to) {
				slog.Error("Invalid from/to format for range type")
				smssender(c, fmt.Errorf("invalid from/to format for range type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "from/to must be YYYY-MM-DD"})
				return
			}

			req.From = from
			req.To = to
		}
	}

	req.UserId = user_id
	req.CourseId = course_id
	req.TariffId = tariff_id
	req.PromoCodeId = promocode_id
	req.Status = status
	req.PaymentType = payment_type
	req.Type = req_type
	req.Day = day
	req.From = from
	req.To = to
	req.Filter = pagination

	slog.Info("Get all orders", "filters", req)
	res, _, err := h.Clients.Order.GetAll(context.Background(), &req)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting all orders: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(200, res)
}
