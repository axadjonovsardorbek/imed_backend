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
// @Summary Create a new lesson
// @Description Create a new lesson with the provided details
// @Tags lesson
// @Accept  json
// @Produce  json
// @Param lesson body models.LessonCreateBody true "Lesson Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /lesson [post]
func (h *Handler) CreateLesson(c *gin.Context) {
	req := m.LessonCreateReq{}
	body := m.LessonCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Name = string(nameJSON)
	req.ModuleId = body.ModuleId
	req.Type = body.Type
	req.OrderNum = body.OrderNum
	req.Duration = body.Duration

	slog.Info("Create lesson", "Lesson", req)
	res, err := h.Clients.Lesson.Create(context.Background(), &req)
	slog.Info("Create lesson response", "Lesson", res)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error creating lesson: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Lesson created successfully")
	c.JSON(200, gin.H{
		"message": "Lesson created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get lesson by ID
// @Description Get a lesson by their ID
// @Tags lesson
// @Accept  json
// @Produce  json
// @Param id path string false "Lesson ID"
// @Success 200 {object} models.LessonRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /lesson/{id} [get]
func (h *Handler) GetByIdLesson(c *gin.Context) {
	id := c.Param("id")

	role := claimData(c, "role")
	user_id := claimData(c, "user_id")

	slog.Info("Get lesson with id", id)
	res, err := h.Clients.Lesson.GetById(context.Background(), &m.CourseById{Id: id, Role: role, UserId: user_id})
	slog.Info("Lesson details", res)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting lesson by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Lesson retrieved successfully")
	c.JSON(200, res)
}

// GetAllLessons godoc
// @Summary Get all lesson
// @Description Get all lesson with optional filtering
// @Tags lesson
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param module_id query string false "Module ID"
// @Param type query string false "Type" Enums(lesson, test)
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.LessonList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /lesson [get]
func (h *Handler) GetAllLessons(c *gin.Context) {
	name := c.Query("name")
	module_id := c.Query("module_id")
	l_type := c.Query("type")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.LessonGetListReq{
		Name:     name,
		ModuleId: module_id,
		Type:     l_type,
		Filter:   pagination,
	}

	slog.Info("Get all lesson with filtes", req)
	data, _, err := h.Clients.Lesson.GetAll(context.Background(), req)
	slog.Info("All lesson", data)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting lesson: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Lessons retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update lesson
// @Description Update lesson
// @Tags lesson
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Lesson ID"
// @Param credentials body models.LessonUpdateBody true "Lesson update credentials"
// @Success 200 {object} string "Lesson updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /lesson/{id}/update [put]
func (h *Handler) UpdateLesson(c *gin.Context) {
	var req m.LessonUpdateReq
	var body m.LessonUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Name = string(nameJSON)
	req.ModuleId = body.ModuleId
	req.Type = body.Type
	req.OrderNum = body.OrderNum
	req.Duration = body.Duration

	slog.Info("Update Lesson", req)
	_, err = h.Clients.Lesson.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error updating Lesson: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson updated"})
}

// Delete godoc
// @Summary Delete lesson
// @Description Delete lesson
// @Tags lesson
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Lesson deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /lesson/{id}/delete [delete]
func (h *Handler) DeleteLesson(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Lesson delete with id", req)
	_, err := h.Clients.Lesson.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error deleting lesson", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
}

// EndLesson godoc
// @Summary End lesson
// @Description End lesson
// @Tags lesson
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Lesson ID"
// @Success 200 {object} string "Lesson ended"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /lesson/{id}/end [put]
func (h *Handler) EndLesson(c *gin.Context) {
	req := m.EndUserLessonReq{}
	id := c.Param("id")

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req.UserId = user_id
	req.LessonId = id

	slog.Info("Update Lesson", req)
	_, err := h.Clients.Lesson.EndLesson(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error updating Lesson: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lesson ended"})
}
