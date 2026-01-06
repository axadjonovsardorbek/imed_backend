package handler

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// @Router /file-upload [post]
// @Success 200 {object} string
func (h *Handler) UploadFile(c *gin.Context) {
	// Faylni olish
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		smssender(c, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := uuid.New().String()

	// Fayl nomini olish
	fileName := id + header.Filename

	// Faylni vaqtinchalik joyga saqlash
	tempFilePath := filepath.Join("/tmp", fileName)
	out, err := os.Create(tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create temporary file"})
		smssender(c, err, http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	// MinIO ga yuklash (avvalgi MinIO.Upload funksiyangizdan foydalanamiz)
	minioURL, err := h.MinIO.Upload(fileName, tempFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to MinIO"})
		smssender(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Successfully upload",
		"Url":     minioURL,
	})
}
