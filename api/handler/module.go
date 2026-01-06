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
// @Summary Create a new module
// @Description Create a new module with the provided details
// @Tags module
// @Accept  json
// @Produce  json
// @Param module body models.ModuleCreateBody true "Module Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /module [post]
func (h *Handler) CreateModule(c *gin.Context) {
	req := m.ModuleCreateReq{}
	body := m.ModuleCreateBody{}
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

	req.Name = string(nameJSON)
	req.CourseId = body.CourseId
	req.OrderNum = body.OrderNum

	slog.Info("Create module", "Module", req)
	res, err := h.Clients.Module.Create(context.Background(), &req)
	slog.Info("Create module response", "Module", res)
	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error creating module: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Module created successfully")
	c.JSON(200, gin.H{
		"Message": "Module created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get module by ID
// @Description Get a module by their ID
// @Tags module
// @Accept  json
// @Produce  json
// @Param id path string false "Module ID"
// @Success 200 {object} models.ModuleRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /module/{id} [get]
func (h *Handler) GetByIdModule(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get module with id", id)
	res, err := h.Clients.Module.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Module details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting module by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Module retrieved successfully")
	c.JSON(200, res)
}

// GetAllModules godoc
// @Summary Get all module
// @Description Get all module with optional filtering
// @Tags module
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param course_id query string false "Course ID"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.ModuleList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /module [get]
func (h *Handler) GetAllModules(c *gin.Context) {
	name := c.Query("name")
	course_id := c.Query("course_id")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.ModuleGetListReq{
		Name:     name,
		CourseId: course_id,
		Filter:   pagination,
	}

	slog.Info("Get all module with filtes", req)
	data, _, err := h.Clients.Module.GetAll(context.Background(), req)
	slog.Info("All module", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting module: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Modules retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update module
// @Description Update module
// @Tags module
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Module ID"
// @Param credentials body models.ModuleUpdateBody true "Module update credentials"
// @Success 200 {object} string "Module updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /module/{id}/update [put]
func (h *Handler) UpdateModule(c *gin.Context) {
	var req m.ModuleUpdateReq
	var body m.ModuleUpdateBody

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

	req.Id = id
	req.Name = string(nameJSON)
	req.CourseId = body.CourseId
	req.OrderNum = body.OrderNum

	slog.Info("Update Module", req)
	_, err = h.Clients.Module.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating Module: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module updated"})
}

// Delete godoc
// @Summary Delete module
// @Description Delete module
// @Tags module
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Module deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /module/{id}/delete [delete]
func (h *Handler) DeleteModule(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Module delete with id", req)
	_, err := h.Clients.Module.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting module", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Module deleted successfully"})
}
