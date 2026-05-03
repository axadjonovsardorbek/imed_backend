package handler

import (
	"imed/pkg/models"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type File struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// File upload
// @Security    BearerAuth
// @Summary File upload
// @Description File upload
// @Tags file-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File"
// @Security BearerAuth
// @Router /web/file-upload [post]
// @Success 200 {object} string
func (h *Handler) UploadFile(c *gin.Context) {

	fileName, tempFilePath, err := formFile(c, "file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		slog.Error("Error processing image: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	} else if fileName == "" && tempFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file not provided"})
		slog.Error("File not provided")
		smssender(c, err, http.StatusBadRequest)
		return
	}

	// MinIO ga yuklash (avvalgi MinIO.Upload funksiyangizdan foydalanamiz)
	minioURL, err := h.MinIO.Upload(fileName, tempFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to upload to MinIO"})
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully upload",
		"Url":     minioURL,
	})
}

// UploadYouTube godoc
// @Security    BearerAuth
// @Summary Upload YouTube video to S3
// @Description Download video from YouTube and upload to S3 storage
// @Tags file-upload
// @Accept json
// @Produce json
// @Param request body models.YouTubeReq true "YouTube URL"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router /web/file-upload-youtube [post]
func (h *Handler) UploadYouTube(c *gin.Context) {
	var req models.YouTubeReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	// core logic
	videoURL, fullID, err := h.S3.UploadFromYouTube(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process youtube video",
		})
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "successfully uploaded",
		"url":        videoURL,
		"youtube_id": fullID,
	})
}

// UploadFileToS3 godoc
// @Security    BearerAuth
// @Summary File upload to S3
// @Description Upload any file directly to S3 storage
// @Tags file-upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Router /web/file-upload-s3 [post]
func (h *Handler) UploadFileToS3(c *gin.Context) {

	fileName, tempFilePath, err := formFile(c, "file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		slog.Error("Error processing image: ", err)
		smssender(c, err, http.StatusBadRequest)
		return
	} else if fileName == "" && tempFilePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file not provided"})
		slog.Error("File not provided")
		smssender(c, err, http.StatusBadRequest)
		return
	}

	// S3 ga yuklash
	s3URL, err := h.S3.Upload(fileName, tempFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to upload to S3"})
		smssender(c, err, http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully upload",
		"Url":     s3URL,
	})
}
