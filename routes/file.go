package routes

import (
	"crypto/subtle"
	"fmt"
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const uploadDirectory = "uploads"

func UploadFile(c *gin.Context) {
	file, err := storeUploadedFile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Dosya başarıyla yüklendi.",
		"file_id":      file.ID,
		"filename":     file.Filename,
		"download_url": downloadURL(file),
	})
}

func DownloadFile(c *gin.Context) {
	var file models.File
	if err := config.DB.First(&file, c.Param("id")).Error; err != nil || file.DownloadToken == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dosya bulunamadı."})
		return
	}
	if subtle.ConstantTimeCompare([]byte(c.Query("token")), []byte(file.DownloadToken)) != 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dosya bulunamadı."})
		return
	}
	if info, err := os.Stat(file.Path); err != nil || !info.Mode().IsRegular() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dosya artık kullanılamıyor."})
		return
	}
	c.FileAttachment(file.Path, file.Filename)
}

func storeUploadedFile(c *gin.Context) (*models.File, error) {
	limitRequestBody(c)
	upload, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("dosya alınamadı")
	}
	if upload.Size <= 0 {
		return nil, fmt.Errorf("boş dosyalar yüklenemez")
	}
	if upload.Size > config.App.MaxUploadBytes {
		return nil, fmt.Errorf("dosya boyutu en fazla %d MB olabilir", config.App.MaxUploadBytes/(1024*1024))
	}

	originalName := filepath.Base(strings.TrimSpace(upload.Filename))
	if originalName == "." || originalName == "" || len(originalName) > 255 {
		return nil, fmt.Errorf("geçersiz dosya adı")
	}
	extension := filepath.Ext(originalName)
	if len(extension) > 16 {
		return nil, fmt.Errorf("geçersiz dosya uzantısı")
	}

	storageToken, err := randomToken(18)
	if err != nil {
		return nil, err
	}
	downloadToken, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(uploadDirectory, 0o750); err != nil {
		return nil, fmt.Errorf("yükleme klasörü oluşturulamadı")
	}
	filePath := filepath.Join(uploadDirectory, storageToken+extension)
	if err := c.SaveUploadedFile(upload, filePath); err != nil {
		return nil, fmt.Errorf("dosya kaydedilemedi")
	}

	user, err := CurrentUser(c)
	if err != nil {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("oturum doğrulanamadı")
	}
	record := &models.File{
		Filename:      originalName,
		Path:          filePath,
		Size:          upload.Size,
		Uploader:      user.Email,
		DownloadToken: downloadToken,
	}
	if err := config.DB.Create(record).Error; err != nil {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("dosya kaydı oluşturulamadı")
	}
	return record, nil
}

func limitRequestBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.App.MaxUploadBytes+1024*1024)
}

func downloadURL(file *models.File) string {
	return fmt.Sprintf("%s/download/%d?token=%s", config.App.AppBaseURL, file.ID, file.DownloadToken)
}
