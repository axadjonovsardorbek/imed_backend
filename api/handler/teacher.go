package handler

import (
	"context"
	"net/http"

	m "imed/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// Create godoc
// @Summary Create a new teacher
// @Description Create a new teacher with the provided details
// @Tags teacher
// @Accept  json
// @Produce  json
// @Param teacher body models.TeacherCreateReq true "Teacher Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /teacher [post]
func (h *Handler) CreateTeacher(c *gin.Context) {
	req := m.TeacherCreateReq{}
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Create teacher", "Teacher", req)
	res, err := h.Clients.Teacher.Create(context.Background(), &req)
	slog.Info("Create teacher response", "Teacher", res)
	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error creating meacher: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Teacher created successfully")
	c.JSON(200, gin.H{
		"Message": "Teacher created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get teacher by ID
// @Description Get a teacher by their ID
// @Tags teacher
// @Accept  json
// @Produce  json
// @Param id path string false "Teacher ID"
// @Success 200 {object} models.TeacherRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /teacher/{id} [get]
func (h *Handler) GetByIdTeacher(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get teacher with id", id)
	res, err := h.Clients.Teacher.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Teacher details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting teacher by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Teacher retrieved successfully")
	c.JSON(200, res)
}

// GetAllTeachers godoc
// @Summary Get all teacher
// @Description Get all teacher with optional filtering
// @Tags teacher
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param phone_number query string false "PhoneNumber"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.TeacherList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /teacher [get]
func (h *Handler) GetAllTeachers(c *gin.Context) {
	name := c.Query("name")
	phone_number := c.Query("phone_number")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.TeacherGetListReq{
		Name:        name,
		PhoneNumber: phone_number,
		Filter:      pagination,
	}

	slog.Info("Get all teacher with filtes", req)
	data, _, err := h.Clients.Teacher.GetAll(context.Background(), req)
	slog.Info("All teacher", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting teacher: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Teachers retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update teacher
// @Description Update teacher
// @Tags teacher
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Teacher ID"
// @Param credentials body models.TeacherUpdateBody true "Teacher update credentials"
// @Success 200 {object} string "Teacher updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /teacher/{id}/update [put]
func (h *Handler) UpdateTeacher(c *gin.Context) {
	var req m.TeacherUpdateReq
	var body m.TeacherUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	req.Id = id
	req.Name = body.Name
	req.PhoneNumber = body.PhoneNumber

	slog.Info("Update Teacher", req)
	_, err := h.Clients.Teacher.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating teacher: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teacher updated"})
}

// Delete godoc
// @Summary Delete teacher
// @Description Delete teacher
// @Tags teacher
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Teacher deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /teacher/{id}/delete [delete]
func (h *Handler) DeleteTeacher(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Teacher delete with id", req)
	_, err := h.Clients.Teacher.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting teacher", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teacher deleted successfully"})
}
