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
// @Summary Create a new subject
// @Description Create a new subject with the provided details
// @Tags subject
// @Accept  json
// @Produce  json
// @Param subject body models.SubjectCreateBody true "Subject Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/subject [post]
func (h *Handler) CreateSubject(c *gin.Context) {
	req := m.SubjectCreateReq{}
	body := m.SubjectCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"message": "failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Name = string(nameJSON)
	req.OrderNum = body.OrderNum
	req.ImageUrl = body.ImageUrl

	slog.Info("Create subject", "Subject", req)
	res, err := h.Clients.Subject.Create(context.Background(), &req)
	slog.Info("Create subject response", "Subject", res)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error creating subject: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Subject created successfully")
	c.JSON(200, gin.H{
		"message": "subject created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get subject by ID
// @Description Get a subject by their ID
// @Tags subject
// @Accept  json
// @Produce  json
// @Param id path string false "Subject ID"
// @Success 200 {object} models.SubjectRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/subject/{id} [get]
func (h *Handler) GetByIdSubject(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get subject with id", id)
	res, err := h.Clients.Subject.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Subject details", res)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting subject by ID: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Subject retrieved successfully")
	c.JSON(200, res)
}

// GetAllSubjects godoc
// @Summary Get all subject
// @Description Get all subject with optional filtering
// @Tags subject
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.SubjectList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/subject [get]
func (h *Handler) GetAllSubjects(c *gin.Context) {
	name := c.Query("name")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.SubjectGetListReq{
		Name:   name,
		Filter: pagination,
	}

	slog.Info("Get all subject with filtes", req)
	data, _, err := h.Clients.Subject.GetAll(context.Background(), req)
	slog.Info("All subject", data)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting subject: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Subjects retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update subject
// @Description Update subject
// @Tags subject
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Subject ID"
// @Param credentials body models.SubjectUpdateBody true "Subject update credentials"
// @Success 200 {object} string "Subject updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /web/subject/{id}/update [put]
func (h *Handler) UpdateSubject(c *gin.Context) {
	var req m.SubjectUpdateReq
	var body m.SubjectUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"message": "failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Name = string(nameJSON)
	req.OrderNum = body.OrderNum
	req.ImageUrl = body.ImageUrl

	slog.Info("Update Subject", req)
	_, err = h.Clients.Subject.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error updating Subject: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subject updated"})
}

// Delete godoc
// @Summary Delete subject
// @Description Delete subject
// @Tags subject
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Subject deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /web/subject/{id}/delete [delete]
func (h *Handler) DeleteSubject(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Subject delete with id", req)
	_, err := h.Clients.Subject.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error deleting subject", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subject deleted successfully"})
}

// GetById godoc
// @Summary Get subject by ID
// @Description Get a subject by their ID
// @Tags subject
// @Accept  json
// @Produce  json
// @Param id path string false "Subject ID"
// @Param Accept-Language header string false "Language (uz|ru|en)" default(uz)
// @Success 200 {object} models.SubjectMobileRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/subject/{id} [get]
func (h *Handler) GetByIdMobileSubject(c *gin.Context) {
	id := c.Param("id")

	lang := getLanguage(c)

	slog.Info("Get subject with id", id)
	res, err := h.Clients.Subject.GetByIdMobile(context.Background(), &m.ByIdMobile{Id: id, Language: lang})
	slog.Info("Subject details", res)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting subject by ID: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Subject retrieved successfully")
	c.JSON(200, res)
}

// GetAllSubjects godoc
// @Summary Get all subject
// @Description Get all subject with optional filtering
// @Tags subject
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param Accept-Language header string false "Language (uz|ru|en)" default(uz)
// @Success 200 {object} models.SubjectMobileList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /mobile/subject [get]
func (h *Handler) GetAllMobileSubjects(c *gin.Context) {
	name := c.Query("name")
	lang := getLanguage(c)

	req := &m.SubjectGetListMobileReq{
		Name:     name,
		Language: lang,
	}

	slog.Info("Get all subject with filtes", req)
	data, _, err := h.Clients.Subject.GetAllMobile(context.Background(), req)
	slog.Info("All subject", data)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting subject: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Subjects retrieved successfully")
	c.JSON(200, data)
}
