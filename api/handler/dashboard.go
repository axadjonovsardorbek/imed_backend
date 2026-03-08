package handler

import (
	"context"
	"fmt"
	m "imed/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

// GetDashboard godoc
// @Summary Get dashboard
// @Description Get dashboard
// @Tags dashboard
// @Accept  json
// @Produce  json
// @Param type query string false "Type (range, day, week, month, year)" Enums(range, day, week, month, year)
// @Param day query string false "Day (YYYY-MM-DD)"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} models.DashboardRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /web/dashboard [get]
func (h *Handler) GetDashboard(c *gin.Context) {
	req := m.GetDashboardReq{}

	req_type := c.Query("type")
	day := c.Query("day")
	from := c.Query("from")
	to := c.Query("to")

	if req_type != "" {
		if req_type == "day" {
			if day != "" && !IsYYYYMMDD(day) {
				slog.Error("Invalid date format for day type")
				smssender(c, fmt.Errorf("invalid date format for day type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "date must be YYYY-MM-DD"})
				return
			}
			req.Day = day
		}

		if req_type == "range" {
			if from == "" || to == "" {
				slog.Error("from and to are required for range type")
				smssender(c, fmt.Errorf("from and to are required for range type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "from and to are required for range"})
				return
			}
			if !IsYYYYMMDD(from) || !IsYYYYMMDD(to) {
				slog.Error("Invalid from/to format for range type")
				smssender(c, fmt.Errorf("invalid from/to format for range type"), http.StatusBadRequest)
				c.JSON(http.StatusBadRequest, gin.H{"message": "from/to must be YYYY-MM-DD"})
				return
			}

			req.From = from
			req.To = to
		}
	}

	req.Type = req_type
	req.Day = day
	req.From = from
	req.To = to

	slog.Info("Get all dashboard with filtes", req)
	data, err := h.Clients.Dashboard.GetDashboard(context.Background(), &req)
	slog.Info("All dashboard", data)

	if err != nil {
		c.JSON(400, gin.H{"message": err.Error()})
		slog.Error("Error getting dashboard: ", err.Error())
		smssender(c, err, http.StatusBadRequest)
		return
	}

	slog.Info("Dashboard retrieved successfully")
	c.JSON(200, data)
}
