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
// @Summary Create a new course
// @Description Create a new course with the provided details
// @Tags course
// @Accept  json
// @Produce  json
// @Param course body models.CourseCreateBody true "Course Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course [post]
func (h *Handler) CreateCourse(c *gin.Context) {
	req := m.CourseCreateReq{}
	body := m.CourseCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	descriptionJson, err := json.Marshal(body.Description)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize description"})
		slog.Error("Error serializing description: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	priceJSON, err := json.Marshal(body.Price)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize price"})
		slog.Error("Error serializing price: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Name = string(nameJSON)
	req.Description = string(descriptionJson)
	req.SubjectId = body.SubjectId
	req.TeacherId = body.TeacherId
	req.Price = string(priceJSON)
	req.ImageUrl = body.ImageUrl
	req.OrderNum = body.OrderNum
	req.IsPublic = body.IsPublic

	slog.Info("Create course", "Course", req)
	res, err := h.Clients.Course.Create(context.Background(), &req)
	slog.Info("Create course response", "Course", res)
	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error creating course: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Course created successfully")
	c.JSON(200, gin.H{
		"Message": "Course created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get course by ID
// @Description Get a course by their ID
// @Tags course
// @Accept  json
// @Produce  json
// @Param id path string false "Course ID"
// @Success 200 {object} models.CourseRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course/{id} [get]
func (h *Handler) GetByIdCourse(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get course with id", id)
	res, err := h.Clients.Course.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Course details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting course by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Course retrieved successfully")
	c.JSON(200, res)
}

// GetAllCourses godoc
// @Summary Get all course
// @Description Get all course with optional filtering
// @Tags course
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param subject_id query string false "Subject ID"
// @Param teacher_id query string false "Teacher ID"
// @Param is_public query bool false "Is Public"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.CourseList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course [get]
func (h *Handler) GetAllCourses(c *gin.Context) {
	name := c.Query("name")
	subject_id := c.Query("subject_id")
	teacher_id := c.Query("teacher_id")
	is_public := c.Query("is_public")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.CourseGetListReq{
		Name:      name,
		SubjectId: subject_id,
		TeacherId: teacher_id,
		IsPublic:  is_public,
		Filter:    pagination,
	}

	slog.Info("Get all course with filtes", req)
	data, _, err := h.Clients.Course.GetAll(context.Background(), req)
	slog.Info("All course", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting course: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Courses retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update course
// @Description Update course
// @Tags course
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Course ID"
// @Param credentials body models.CourseUpdateBody true "Course update credentials"
// @Success 200 {object} string "Course updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /course/{id}/update [put]
func (h *Handler) UpdateCourse(c *gin.Context) {
	var req m.CourseUpdateReq
	var body m.CourseUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	descriptionJson, err := json.Marshal(body.Description)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize description"})
		slog.Error("Error serializing description: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	nameJSON, err := json.Marshal(body.Name)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize name"})
		slog.Error("Error serializing name: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}
	priceJSON, err := json.Marshal(body.Price)
	if err != nil {
		c.JSON(409, gin.H{"Error": "Failed to serialize price"})
		slog.Error("Error serializing price: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Name = string(nameJSON)
	req.Description = string(descriptionJson)
	req.SubjectId = body.SubjectId
	req.TeacherId = body.TeacherId
	req.Price = string(priceJSON)
	req.ImageUrl = body.ImageUrl
	req.OrderNum = body.OrderNum
	req.IsPublic = body.IsPublic

	slog.Info("Update Course", req)
	_, err = h.Clients.Course.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating Course: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course updated"})
}

// Delete godoc
// @Summary Delete course
// @Description Delete course
// @Tags course
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Course deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /course/{id}/delete [delete]
func (h *Handler) DeleteCourse(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Course delete with id", req)
	_, err := h.Clients.Course.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting course", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course deleted successfully"})
}

// Buy godoc
// @Summary Buy a course
// @Description Buy a course with the provided details
// @Tags course
// @Accept  json
// @Produce  json
// @Param course body models.UserCourseCreateReq true "Course Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course/permission [post]
func (h *Handler) BuyCourse(c *gin.Context) {
	req := m.UserCourseCreateReq{}
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Buy course", "Course", req)
	res, err := h.Clients.Course.BuyCourse(context.Background(), &req)
	slog.Info("Buy course response", "Course", res)
	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error buying course: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Course bought successfully")
	c.JSON(200, gin.H{
		"Message": "Course bought successfully",
	})
}

// GetById godoc
// @Summary Get course by ID
// @Description Get a course by their ID
// @Tags course
// @Accept  json
// @Produce  json
// @Param id path string false "Course ID"
// @Success 200 {object} models.CourseRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course/permission/{id} [get]
func (h *Handler) UserCourse(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get user course with id", id)
	res, err := h.Clients.Course.UserCourse(context.Background(), &m.ById{Id: id})
	slog.Info("User course details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting user course by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User course retrieved successfully")
	c.JSON(200, res)
}

// GetAllUserCourses godoc
// @Summary Get all user course
// @Description Get all user course with optional filtering
// @Tags course
// @Accept  json
// @Produce  json
// @Param type query string false "Type"
// @Param is_active query string false "Is Active"
// @Param user_id query string false "User ID"
// @Param course_id query string false "Course ID"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.UserCourseList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /course/permission [get]
func (h *Handler) UserCourses(c *gin.Context) {
	c_type := c.Query("type")
	user_id := c.Query("user_id")
	course_id := c.Query("course_id")
	is_active := c.Query("is_active")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.UserCourseGetListReq{
		Type:     c_type,
		UserId:   user_id,
		CourseId: course_id,
		IsActive: is_active,
		Filter:   pagination,
	}

	slog.Info("Get all user course with filtes", req)
	data, _, err := h.Clients.Course.UserCourses(context.Background(), req)
	slog.Info("All user course", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting user course: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User courses retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update user course
// @Description Update user course
// @Tags course
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Permission ID"
// @Param credentials body models.UserCourseUpdateReq true "Course update credentials"
// @Success 200 {object} string "Permission updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /course/permission{id}/update [put]
func (h *Handler) UpdateUserCourse(c *gin.Context) {
	var req m.UserCourseUpdateReq

	id := c.Param("id")

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	req.Id = id

	slog.Info("Update user uourse", req)
	_, err := h.Clients.Course.UserUpdateCourse(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating user Course: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User course updated"})
}

// Delete godoc
// @Summary Delete course
// @Description Delete course
// @Tags course
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Course deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /course/permission/{id}/delete [delete]
func (h *Handler) DeleteUserCourse(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("User course delete with id", req)
	_, err := h.Clients.Course.DeleteUserCourse(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting user course", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User course deleted successfully"})
}
