package handlers

import (
	"crypto/md5"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-fast-cdn/database"
	"github.com/google/uuid"
)

func HandleDocsUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("doc")

	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read file: %s", err.Error())
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to open file: %s", err.Error())
		return
	}
	defer file.Close()

	fileBuffer := make([]byte, 512)
	_, err = file.Read(fileBuffer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read file: %s", err.Error())
		return
	}
	fileType := http.DetectContentType(fileBuffer)

	allowedMimeTypes := map[string]bool{
		"text/plain":         true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/pdf":       true,
		"application/rtf":       true,
		"application/x-freearc": true,
	}

	if !allowedMimeTypes[fileType] {
		c.String(http.StatusBadRequest, "Invalid file type: %s", fileType)
		return
	}

	fileHashBuffer := md5.Sum(fileBuffer)
	fileName := uuid.NewString() + filepath.Ext(fileHeader.Filename)
	savedFileName, alreadyExists := database.AddDoc(fileName, fileHashBuffer[:])

	if !alreadyExists {
		err = c.SaveUploadedFile(fileHeader, "./uploads/docs/"+fileName)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
			return
		}
	}

	body := gin.H{
		"file_url": "localhost:8080/" + "download/docs/" + savedFileName,
	}

	c.JSON(http.StatusOK, body)
}
