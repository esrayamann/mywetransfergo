package routes

import (
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Admin sayfası
func AdminPage(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcılar yüklenemedi"})
		return
	}

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"users": users,
	})
}

// Rol değiştirme
func ChangeUserRole(c *gin.Context) {
	id := c.PostForm("id")

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	if user.Role == "admin" {
		user.Role = "user"
	} else {
		user.Role = "admin"
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rol güncellenemedi"})
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}
