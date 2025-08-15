package routes

import (
	"fmt"
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	// Formdan dosya al
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dosya yüklenemedi"})
		return
	}

	// uploads klasörü yoksa oluştur
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Uploads klasörü oluşturulamadı"})
			return
		}
	}

	// Benzersiz dosya adı oluştur
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Dosyayı kaydet
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Dosya kaydedilemedi"})
		return
	}

	// DB'ye kaydet
	newFile := models.File{
		Filename: file.Filename,
		Path:     filePath,
		Size:     file.Size,
		Uploader: "anonymous", // giriş yapmış kullanıcı varsa onun emaili/username buraya gelir
	}

	if err := config.DB.Create(&newFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanına eklenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Dosya başarıyla yüklendi",
		"file_id":  newFile.ID,
		"filename": newFile.Filename,
	})
}

func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	var file models.File

	if err := config.DB.First(&file, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dosya bulunamadı"})
		return
	}

	// Dosya indirme header'ı ekle
	c.Header("Content-Disposition", "attachment; filename="+file.Filename)
	c.Header("Content-Type", "application/octet-stream")

	c.File(file.Path)
}
