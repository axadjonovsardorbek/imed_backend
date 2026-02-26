package handler

import (
	"context"
	"net/http"

	m "imed/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// GetById godoc
// @Summary Get admin by ID
// @Description Get a admin by their ID
// @Tags admin
// @Accept  json
// @Produce  json
// @Param id path string false "Admin ID"
// @Success 200 {object} models.AdminRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/admin/{id} [get]
func (h *Handler) GetByIdAdmin(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get admin with id", id)
	res, err := h.Clients.Admin.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("Admin details", res)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting admin by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("admin retrieved successfully")
	c.JSON(200, res)
}

// Login godoc
// @Summary Login admin
// @Description Authenticate admin with login and password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.LoginReq true "Admin login credentials"
// @Success 200 {object} models.TokenRes "JWT tokens"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /web/auth/admin/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req m.LoginReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Admin login credentials ", req)
	res, err := h.Clients.Admin.Login(context.Background(), &req)
	slog.Info("Admin login response", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error while login", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}