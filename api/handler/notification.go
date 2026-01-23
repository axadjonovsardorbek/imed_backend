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
// @Summary Create a new notification
// @Description Create a new notification with the provided details
// @Tags notification
// @Accept  json
// @Produce  json
// @Param notification body models.NotificationCreateBody true "Notification Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /notification [post]
func (h *Handler) CreateNotification(c *gin.Context) {
	req := m.NotificationCreateReq{}
	body := m.NotificationCreateBody{}
	err := c.BindJSON(&body)
	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error binding request body: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	}

	titleJSON, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize title"})
		slog.Error("Error serializing title: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	messageJSON, err := json.Marshal(body.Message)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize message"})
		slog.Error("Error serializing message: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Title = string(titleJSON)
	req.Message = string(messageJSON)
	req.Type = body.Type
	req.CourseId = body.CourseId

	slog.Info("Create notification", "Notification", req)
	res, err := h.Clients.Notification.Create(context.Background(), &req)
	slog.Info("Create notification response", "Notification", res)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error creating notification: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Notification created successfully")
	c.JSON(200, gin.H{
		"message": "Notification created successfully",
		"Id":      res.Id,
	})
}

// GetById godoc
// @Summary Get notification by ID
// @Description Get a notification by their ID
// @Tags notification
// @Accept  json
// @Produce  json
// @Param id path string false "Notification ID"
// @Success 200 {object} models.MasterNotificationRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /notification/{id} [get]
func (h *Handler) GetByIdNotification(c *gin.Context) {
	id := c.Param("id")

	slog.Info("Get notification with id", id)
	res, err := h.Clients.Notification.GetById(context.Background(), &m.ById{Id: id})
	slog.Info("otification details", res)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting notification by ID: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Notification retrieved successfully")
	c.JSON(200, res)
}

// GetAllNotifications godoc
// @Summary Get all notification
// @Description Get all notification with optional filtering
// @Tags notification
// @Accept  json
// @Produce  json
// @Param title query string false "Title"
// @Param course_id query string false "Course ID"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.MasterNotificationList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /notification [get]
func (h *Handler) GetAllNotifications(c *gin.Context) {
	title := c.Query("title")
	course_id := c.Query("course_id")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	req := &m.NotificationGetListReq{
		Title:    title,
		CourseId: course_id,
		Filter:   pagination,
	}

	slog.Info("Get all notification with filtes", req)
	data, _, err := h.Clients.Notification.GetAll(context.Background(), req)
	slog.Info("All notification", data)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting notification: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Notifications retrieved successfully")
	c.JSON(200, data)
}

// Update godoc
// @Summary Update notification
// @Description Update notification
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Notification ID"
// @Param credentials body models.MasterNotificationUpdateBody true "Notification update credentials"
// @Success 200 {object} string "Notification updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /notification/{id}/update [put]
func (h *Handler) UpdateNotification(c *gin.Context) {
	var req m.MasterNotificationUpdateReq
	var body m.MasterNotificationUpdateBody

	id := c.Param("id")

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		slog.Error("Error binding request body", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	titleJSON, err := json.Marshal(body.Title)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize title"})
		slog.Error("Error serializing title: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	messageJSON, err := json.Marshal(body.Message)
	if err != nil {
		c.JSON(409, gin.H{"message": "Failed to serialize message"})
		slog.Error("Error serializing message: ", err)
		smssender(c, err, http.StatusConflict)
		return
	}

	req.Id = id
	req.Title = string(titleJSON)
	req.Message = string(messageJSON)

	slog.Info("Update Notification", req)
	_, err = h.Clients.Notification.Update(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error updating Notification: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification updated"})
}

// Delete godoc
// @Summary Delete notification
// @Description Delete notification
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Id"
// @Success 200 {object} string "Notification deleted"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /notification/{id}/delete [delete]
func (h *Handler) DeleteNotification(c *gin.Context) {
	var req m.ById

	id := c.Param("id")

	req.Id = id

	slog.Info("Notification delete with id", req)
	_, err := h.Clients.Notification.Delete(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error deleting notification", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetAllNotifications godoc
// @Summary Get all notification
// @Description Get all notification with optional filtering
// @Tags notification
// @Accept  json
// @Produce  json
// @Param is_read query bool false "Is Read"
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Success 200 {object} models.UserNotificationList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /notification/user [get]
func (h *Handler) GetAllUserNotifications(c *gin.Context) {
	is_read := c.Query("is_read")
	limit := c.Query("limit")
	page := c.Query("page")

	pagination := m.ParsePagination(limit, page)

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req := &m.NotificationGetListReq{
		UserId: user_id,
		IsRead: is_read,
		Filter: pagination,
	}

	slog.Info("Get all notification with filtes", req)
	data, _, err := h.Clients.Notification.UserNotifications(context.Background(), req)
	slog.Info("All notification", data)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		slog.Error("Error getting notification: ", err.Error())
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	slog.Info("Notifications retrieved successfully")
	c.JSON(200, data)
}

// Read godoc
// @Summary Read notification
// @Description Read notification
// @Tags notification
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string false "Notification ID"
// @Success 200 {object} string "Notification updated"
// @Failure 400 {object} string "Invalid request payload"
// @Failure 500 {object} string "Server error"
// @Router /notification/{id}/read [put]
func (h *Handler) ReadNotification(c *gin.Context) {
	var req m.NotificationMarkReadReq

	id := c.Param("id")

	user_id := claimData(c, "user_id")
	if user_id == "" {
		return
	}

	req.Id = id
	req.UserId = user_id

	slog.Info("Update Notification", req)
	_, err := h.Clients.Notification.MarkAsRead(context.Background(), &req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		slog.Error("Error updating Notification: ", err)
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification read"})
}
