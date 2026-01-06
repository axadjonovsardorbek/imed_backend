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
// @Summary Create a new source
// @Description Create a new source with the provided details
// @Tags source
// @Accept  json
// @Produce  json
// @Param source body models.SourceCreateBody true "Source Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /source [post]
func (h *Handler) CreateSource(c *gin.Context) {
	req := m.SourceCreateReq{}
	body := m.SourceCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	urlJSON, err := json.Marshal(body.Url)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize url"})
		slog.Error("Error serializing url: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Name = string(nameJSON)
	req.LessonId = body.LessonId
	req.Type = body.Type
	req.OrderNum = body.OrderNum
	req.Url = string(urlJSON)

	slog.Info("Create source", "Source", req)
	res, err := h.Clients.Source.Create(context.Background(), &req)
	slog.Info("Create source response", "Source", res)
	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error creating source: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Source created successfully")
	c.JSON(200, gin.H{
		"Message": "Source created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get source by ID
// @Description Get a source by their ID
// @Tags source
// @Accept  json
// @Produce  json
// @Param id path string false "Source ID"
// @Success 200 {object} models.SourceRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /source/{id} [get]
func (h *Handler) GetByIdSource(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get source with id", id)
	res, err := h.Clients.Source.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Source details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting source by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Source retrieved successfully")
	c.JSON(200, res)
}

// GetAllSources godoc
// @Summary Get all source
// @Description Get all source with optional filtering
// @Tags source
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param lesson_id query string false "Lesson ID"
// @Param type query string false "Type" Enums(video, document, test)
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.SourceList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /source [get]
func (h *Handler) GetAllSources(c *gin.Context) {
	name := c.Query("name")
	lesson_id := c.Query("lesson_id")
	s_type := c.Query("type")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.SourceGetListReq{
		Name:     name,
		LessonId: lesson_id,
		Type:     s_type,
		Filter:   pagination,
	}

	slog.Info("Get all source with filtes", req)
	data, _, err := h.Clients.Source.GetAll(context.Background(), req)
	slog.Info("All source", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting source: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Sources retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update source
// @Description Update source
// @Tags source
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Source ID"
// @Param credentials body models.SourceUpdateBody true "Source update credentials"
// @Success 200 {object} string "Source updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /source/{id}/update [put]
func (h *Handler) UpdateSource(c *gin.Context) {
	var req m.SourceUpdateReq
	var body m.SourceUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	urlJSON, err := json.Marshal(body.Url)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize url"})
		slog.Error("Error serializing url: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Name = string(nameJSON)
	req.LessonId = body.LessonId
	req.Type = body.Type
	req.OrderNum = body.OrderNum
	req.Url = string(urlJSON)

	slog.Info("Update Source", req)
	_, err = h.Clients.Source.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating Source: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Source updated"})
}

// Delete godoc
// @Summary Delete source
// @Description Delete source
// @Tags source
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Source deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /source/{id}/delete [delete]
func (h *Handler) DeleteSource(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Source delete with id", req)
	_, err := h.Clients.Source.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting source", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Source deleted successfully"})
}
