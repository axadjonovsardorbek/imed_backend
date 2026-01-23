package handler

import (
	"context"
	"encoding/json"
	"net/http"

	m "imed/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// Create godoc
// @Summary Create a new banner
// @Description Create a new banner with the provided details
// @Tags banner
// @Accept  json
// @Produce  json
// @Param banner body models.BannerCreateBody true "Banner Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /banner [post]
func (h *Handler) CreateBanner(c *gin.Context) {
	req := m.BannerCreateReq{}
	body := m.BannerCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	titleJson, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize title"})
		slog.Error("Error serializing title: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	descriptionJson, err := json.Marshal(body.Description)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize description"})
		slog.Error("Error serializing description: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	imageurlJson, err := json.Marshal(body.ImageUrl)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize image"})
		slog.Error("Error serializing image: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Title = string(titleJson)
	req.Description = string(descriptionJson)
	req.ImageUrl = string(imageurlJson)
	req.LinkUrl = body.LinkUrl
	req.OrderNum = body.OrderNum

	slog.Info("Create banner", "Banner", req)
	res, err := h.Clients.Banner.Create(context.Background(), &req)
	slog.Info("Create banner response", "Banner", res)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error creating banner: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Banner created successfully")
	c.JSON(200, gin.H{
		"message": "Banner created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get banner by ID
// @Description Get a banner by their ID
// @Tags banner
// @Accept  json
// @Produce  json
// @Param id path string false "Banner ID"
// @Success 200 {object} models.BannerRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /banner/{id} [get]
func (h *Handler) GetByIdBanner(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get banner with id", id)
	res, err := h.Clients.Banner.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Banner details", res)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting banner by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Banner retrieved successfully")
	c.JSON(200, res)
}

// GetAllBanners godoc
// @Summary Get all banner
// @Description Get all banner with optional filtering
// @Tags banner
// @Accept  json
// @Produce  json
// @Param title query string false "Title"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.BannerList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /banner [get]
func (h *Handler) GetAllBanners(c *gin.Context) {
	title := c.Query("title")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.BannerGetListReq{
		Title:  title,
		Filter: pagination,
	}

	slog.Info("Get all banner with filtes", req)
	data, _, err := h.Clients.Banner.GetAll(context.Background(), req)
	slog.Info("All banner", data)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting banner: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Banners retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update banner
// @Description Update banner
// @Tags banner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Banner ID"
// @Param credentials body models.BannerUpdateBody true "Banner update credentials"
// @Success 200 {object} string "Banner updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /banner/{id}/update [put]
func (h *Handler) UpdateBanner(c *gin.Context) {
	var req m.BannerUpdateReq
	var body m.BannerUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	titleJson, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize title"})
		slog.Error("Error serializing title: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	descriptionJson, err := json.Marshal(body.Description)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize description"})
		slog.Error("Error serializing description: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	imageurlJson, err := json.Marshal(body.ImageUrl)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize image"})
		slog.Error("Error serializing image: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Title = string(titleJson)
	req.Description = string(descriptionJson)
	req.ImageUrl = string(imageurlJson)
	req.LinkUrl = body.LinkUrl
	req.OrderNum = body.OrderNum

	slog.Info("Update Banner", req)
	_, err = h.Clients.Banner.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error updating Banner: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Banner updated"})
}

// Delete godoc
// @Summary Delete banner
// @Description Delete banner
// @Tags banner
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Banner deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /banner/{id}/delete [delete]
func (h *Handler) DeleteBanner(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Banner delete with id", req)
	_, err := h.Clients.Banner.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error deleting banner", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Banner deleted successfully"})
}
