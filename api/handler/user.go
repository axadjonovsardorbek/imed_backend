package handler

import (
	"context"
	"fmt"
	"net/http"

	m "imed/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// ChangePassword godoc
// @Summary Change password
// @Description Change password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.ChangePasswordBody true "Change password credentials"
// @Success 200 {object} string "Password changed successfully"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /auth/password/change [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req m.ChangePasswordReq
	var body m.ChangePasswordBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	if body.OldPassword == body.NewPassword || body.OldPassword == body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "The old password must not be the same as the new password.",
		})
		slog.Error("Error changing password")
		smssender(c, fmt.Errorf("the old password same as the new password"), http.StatusBadRequest)
		return
	}
	if body.NewPassword != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "The new password must not be the different as the confirm password.",
		})
		slog.Error("Error changing password")
		smssender(c, fmt.Errorf("the confirm password different as the new password"), http.StatusBadRequest)
		return
	}

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}
	role := claimData(c, "role")
	if role == "" {
		return
	}

	req.NewPassword = body.NewPassword
	req.OldPassword = body.OldPassword
	req.Id = user_id
	req.Role = role

	slog.Info("Change password credentials ", req)
	res, err := h.Clients.User.ChangePassword(context.Background(), &req)
	slog.Info("Change password response", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error while check", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}

// RefreshPassword godoc
// @Summary Refresh password
// @Description Refresh password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.RefreshPasswordReq true "Refresh password credentials"
// @Success 200 {object} models.RefreshPassword
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /auth/password/refresh [put]
func (h *Handler) RefreshPassword(c *gin.Context) {

	var req m.RefreshPasswordReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Refresh password", req.UserId)
	res, err := h.Clients.User.RefreshPassword(context.Background(), &req)
	slog.Info("Refreshed password", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error refreshing password: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Check godoc
// @Summary Check user
// @Description Check user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.UserCheckReq true "User check credentials"
// @Success 200 {object} bool "User exists or not"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /auth/user/check [post]
func (h *Handler) CheckUser(c *gin.Context) {
	var req m.UserCheckReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("User check credentials ", req)
	res, err := h.Clients.User.Check(context.Background(), &req)
	slog.Info("User check response", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error while check", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with login and password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.UserLoginReq true "User login credentials"
// @Success 200 {object} models.TokenRes "JWT tokens"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /auth/user/login [post]
func (h *Handler) LoginUser(c *gin.Context) {
	var req m.UserLoginReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("User login credentials ", req)
	res, err := h.Clients.User.Login(context.Background(), &req)
	slog.Info("User login response", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error while login", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetById godoc
// @Summary Get user by ID
// @Description Get an user by their ID
// @Tags user
// @Accept  json
// @Produce  json
// @Param id path string false "User ID"
// @Success 200 {object} models.UserRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /user/{id} [get]
func (h *Handler) GetByIdUser(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get user with id", id)
	res, err := h.Clients.User.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("User details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting user by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User retrieved successfully")
	c.JSON(200, res)
}

// GetAllUsers godoc
// @Summary Get all user
// @Description Get all user with optional filtering
// @Tags user
// @Accept  json
// @Produce  json
// @Param name query string false "Name"
// @Param phone_number query string false "PhoneNumber"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.UserList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /user [get]
func (h *Handler) GetAllUsers(c *gin.Context) {
	name := c.Query("name")
	phone_number := c.Query("phone_number")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.UserGetListReq{
		Name:        name,
		PhoneNumber: phone_number,
		Filter:      pagination,
	}

	slog.Info("Get all user with filtes", req)
	data, _, err := h.Clients.User.GetAll(context.Background(), req)
	slog.Info("All user", data)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting user: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Users retrieved successfully")
	c.JSON(200, data)
}

// Delete godoc
// @Summary Delete user
// @Description Delete user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "User deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /user/{id}/delete [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("User delete with id", req)
	_, err := h.Clients.User.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting user", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get user profile
// @Tags profile
// @Accept  json
// @Produce  json
// @Success 200 {object} models.UserRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /user/get/profile [get]
func (h *Handler) GetProfileUser(c *gin.Context) {

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	slog.Info("Get user with id", user_id)
	res, err := h.Clients.User.GetById(context.Background(), &m.ById{Id: user_id})
	slog.Info("User details", res)

	if err != nil {
		c.JSON(500, gin.H{"Error": err.Error()})
		slog.Error("Error getting user by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("User retrieved successfully")
	c.JSON(200, res)
}

// Update godoc
// @Summary Update a user
// @Description Update a user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.UserUpdateBody true "User update credentials"
// @Success 200 {object} string "Profile updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /user/update/profile [put]
func (h *Handler) UpdateProfileUser(c *gin.Context) {
	var req m.UserUpdateReq
	var body m.UserUpdateBody

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	req.Id = user_id
	req.Name = body.Name
	req.ImageUrl = body.ImageUrl
	req.PhoneNumber = body.PhoneNumber
	req.FCMToken = body.FCMToken
	req.Language = body.Language

	slog.Info("Update User Profile", req)
	_, err := h.Clients.User.Update(context.Background(), &req)
	slog.Info("User profile updated")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error updating user profile: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

// Delete godoc
// @Summary Delete a profile
// @Description Delete a profile
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} string "Profile deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /user/delete/profile [delete]
func (h *Handler) DeleteProfileUser(c *gin.Context) {
	var req m.ById

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req.Id = user_id

	slog.Info("User delete profile with id", req)
	_, err := h.Clients.User.Delete(context.Background(), &req)
	slog.Info("User profile deleted")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error deleting user profile", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

// Activity godoc
// @Summary Activity user
// @Description Activity user
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param credentials body models.UserActivityCreateBody true "User activity credentials"
// @Success 200 {object} string "Activity created"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /user/activity [post]
func (h *Handler) CreateActivity(c *gin.Context) {
	var req m.UserActivityCreateReq
	var body m.UserActivityCreateBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Error("Error binding request body", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req.UserId = user_id
	req.Activity = body.Activity

	slog.Info("Create activity ", req)
	res, err := h.Clients.User.CreateActivity(context.Background(), &req)
	slog.Info("Create activity response", res)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("Error while create activity", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, res)
}
