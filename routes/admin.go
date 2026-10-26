package routes

import (
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AdminPage(c *gin.Context) {
	currentUser, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcılar yüklenemedi."})
		return
	}
	c.HTML(http.StatusOK, "admin.html", gin.H{
		"Username":  currentUser.Username,
		"CSRFToken": csrfToken(c),
		"users":     users,
	})
}

func ChangeUserRole(c *gin.Context) {
	currentUser, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	id, err := strconv.ParseUint(c.PostForm("id"), 10, 64)
	if err != nil || id == 0 || uint(id) == currentUser.ID {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var user models.User
	if err := config.DB.First(&user, uint(id)).Error; err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if user.Role == "admin" {
		user.Role = "user"
	} else {
		user.Role = "admin"
	}
	if err := config.DB.Save(&user).Error; err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin")
}
