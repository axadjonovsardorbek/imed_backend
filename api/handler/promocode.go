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
// @Summary Create a new promocode
// @Description Create a new promocode with the provided details
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param promocode body models.PromoCodeCreateReq true "PromoCode Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/promocode [post]
func (h *Handler) CreatePromoCode(c *gin.Context) {
	req := m.PromoCodeCreateReq{}

	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	if req.Type == "string" || req.Type == "" {
		req.Type = "all"
	}

	if req.Type != "all" && req.Type != "selected" {
		c.JSON(400, gin.H{"message": "incorrect type for promocode, all or selected for type"})
		slog.Error("incorrect type for promocode, all or selected for type")
		smssender(c, fmt.Errorf("incorrect type for promocode, all or selected for type"), http.StatusBadRequest)
		return
	}

	if req.DiscountType != "fixed" && req.DiscountType != "percent" {
		c.JSON(400, gin.H{"message": "incorrect type for discount, fixed or percent for discout type"})
		slog.Error("incorrect type for discount, fixed or percent for discout type")
		smssender(c, fmt.Errorf("incorrect type for discount, fixed or percent for discout type"), http.StatusBadRequest)
		return
	}

	if !IsYYYYMMDD(req.StartsAt) || !IsYYYYMMDD(req.EndsAt) {
		slog.Error("Invalid start/end date")
		smssender(c, fmt.Errorf("invalid start/end date"), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"message": "start/end date must be YYYY-MM-DD"})
		return
	}

	if req.Type == "selected" && len(req.Courses) == 0 {
		msg := "courses are required when type is 'selected'"

		c.JSON(http.StatusBadRequest, gin.H{"message": msg})
		slog.Error(msg)
		smssender(c, fmt.Errorf(msg), http.StatusBadRequest)
		return
	}

	if req.DiscountType == "fixed" {
		req.MaxDiscount = req.DiscountValue
	}

	slog.Info("Create promocode", "PromoCode", req)
	res, err := h.Clients.PromoCode.Create(context.Background(), &req)
	slog.Info("Create promocode response", "PromoCode", res)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error creating promocode: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCode created successfully")
	c.JSON(200, gin.H{
		"message": "promocode created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get promocode by ID
// @Description Get a promocode by their ID
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param id path string false "PromoCode ID"
// @Success 200 {object} models.PromoCodeRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/promocode/{id} [get]
func (h *Handler) GetByIdPromoCode(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get promocode with id", id)
	res, err := h.Clients.PromoCode.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("PromoCode details", res)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting promocode by ID: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCode retrieved successfully")
	c.JSON(200, res)
}

// GetAllPromoCodes godoc
// @Summary Get all promocode
// @Description Get all promocode with optional filtering
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param code query string false "Code"
// @Param discount_type query string false "Discount Type"
// @Param is_active query string false "IsActive" Enums(true, false)
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.PromoCodeList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/promocode [get]
func (h *Handler) GetAllPromoCodes(c *gin.Context) {
	code := c.Query("code")
	discount_type := c.Query("discount_type")
	is_active := c.Query("is_active")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.PromoCodeGetListReq{
		Code:         code,
		DiscountType: discount_type,
		IsActive:     is_active,
		Filter:       pagination,
	}

	slog.Info("Get all promocode with filtes", req)
	data, _, err := h.Clients.PromoCode.GetAll(context.Background(), req)
	slog.Info("All promocode", data)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting promocode: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCodes retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update promocode
// @Description Update promocode
// @Tags promocode
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "PromoCode ID"
// @Param credentials body models.PromoCodeUpdateBody true "PromoCode update credentials"
// @Success 200 {object} string "PromoCode updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /web/promocode/{id}/update [put]
func (h *Handler) UpdatePromoCode(c *gin.Context) {
	var req m.PromoCodeUpdateReq
	var body m.PromoCodeUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	req.Id = id
	req.DiscountType = body.DiscountType
	req.DiscountValue = body.DiscountValue
	req.EndsAt = body.EndsAt
	req.StartsAt = body.StartsAt
	req.IsActive = body.IsActive
	req.MaxDiscount = body.MaxDiscount
	req.MaxUsesPerUser = body.MaxUsesPerUser
	req.MaxUsesTotal = body.MaxUsesTotal
	req.MinOrderAmount = body.MinOrderAmount

	if req.Type == "string" || req.Type == "" {
		req.Type = "all"
	}

	if req.Type != "all" && req.Type != "selected" {
		c.JSON(400, gin.H{"message": "incorrect type for promocode, all or selected for type"})
		slog.Error("incorrect type for promocode, all or selected for type")
		smssender(c, fmt.Errorf("incorrect type for promocode, all or selected for type"), http.StatusBadRequest)
		return
	}

	if req.DiscountType != "fixed" && req.DiscountType != "percent" {
		c.JSON(400, gin.H{"message": "incorrect type for discount, fixed or percent for discout type"})
		slog.Error("incorrect type for discount, fixed or percent for discout type")
		smssender(c, fmt.Errorf("incorrect type for discount, fixed or percent for discout type"), http.StatusBadRequest)
		return
	}

	if !IsYYYYMMDD(req.StartsAt) || !IsYYYYMMDD(req.EndsAt) {
		slog.Error("Invalid start/end date")
		smssender(c, fmt.Errorf("invalid start/end date"), http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"message": "start/end date must be YYYY-MM-DD"})
		return
	}

	if req.Type == "selected" && len(req.Courses) == 0 {
		msg := "courses are required when type is 'selected'"

		c.JSON(http.StatusBadRequest, gin.H{"message": msg})
		slog.Error(msg)
		smssender(c, fmt.Errorf(msg), http.StatusBadRequest)
		return
	}

	if req.DiscountType == "fixed" {
		req.MaxDiscount = req.DiscountValue
	}

	slog.Info("Update PromoCode", req)
	_, err := h.Clients.PromoCode.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error updating PromoCode: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PromoCode updated"})
}

// Delete godoc
// @Summary Delete promocode
// @Description Delete promocode
// @Tags promocode
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "PromoCode deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /web/promocode/{id}/delete [delete]
func (h *Handler) DeletePromoCode(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("PromoCode delete with id", req)
	_, err := h.Clients.PromoCode.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error deleting promocode", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promocode deleted successfully"})
}

// CheckPromoCode godoc
// @Summary Check promocode
// @Description Check a promocode with the provided details
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param promocode body models.PromoCodeCheckBody true "PromoCode Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/promocode/check [post]
func (h *Handler) CheckPromoCode(c *gin.Context) {
	body := m.PromoCodeCheckBody{}
	req := m.PromoCodeCheckReq{}

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

	req.PromoCode = body.PromoCode
	req.UserId = user_id
	req.CourseId = body.CourseId
	req.TariffId = body.TariffId

	slog.Info("Check promocode", "PromoCode", req)
	res, err := h.Clients.PromoCode.Check(context.Background(), &req)
	slog.Info("Check promocode response", "PromoCode", res)
	if err != nil {
		c.JSON(400, res)
		slog.Error("Error checking promocode: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCode created successfully")
	c.JSON(200, res)
}

// GetById godoc
// @Summary Get promocode by ID
// @Description Get a promocode by their ID
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param id path string false "PromoCodeRedemption ID"
// @Success 200 {object} models.PromoRedemptionRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/promocode/redemption/{id} [get]
func (h *Handler) GetByIdUserPromoCode(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get promocode with id", id)
	res, err := h.Clients.PromoCode.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("PromoCode details", res)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting promocode by ID: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCode retrieved successfully")
	c.JSON(200, res)
}

// GetAllPromoCodes godoc
// @Summary Get all promocode
// @Description Get all promocode with optional filtering
// @Tags promocode
// @Accept  json
// @Produce  json
// @Param user_id query string false "UserId"
// @Param promo_id query string false "PromoCode ID"
// @Param is_active query string false "IsActive" Enums(true, false)
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.PromoRedemptionList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/promocode/redemption [get]
func (h *Handler) GetAllUserPromoCodes(c *gin.Context) {
	user_id := c.Query("user_id")
	promo_id := c.Query("promo_id")
	is_active := c.Query("is_active")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.PromoRedemptionGetListReq{
		UserId:   user_id,
		PromoId:  promo_id,
		IsActive: is_active,
		Filter:   pagination,
	}

	slog.Info("Get all promocode with filtes", req)
	data, _, err := h.Clients.PromoCode.GetAllUserPromo(context.Background(), req)
	slog.Info("All promocode", data)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting promocode: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("PromoCodeRedemptions retrieved successfully")
	c.JSON(200, data)
}
