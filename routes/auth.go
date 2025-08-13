package routes

import (
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GET /login
func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Error":    "",
		"Username": "",
	})
}

// POST /login
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user models.User
	if err := config.DB.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Error":    "Kullanıcı bulunamadı veya şifre yanlış",
			"Username": username,
		})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Error":    "Kullanıcı bulunamadı veya şifre yanlış",
			"Username": username,
		})
		return
	}

	// Başarılı giriş için cookie ayarlıyoruz
	c.SetCookie("user_id", strconv.FormatUint(uint64(user.ID), 10), 3600, "/", "", false, true)
	c.SetCookie("role", user.Role, 3600, "/", "", false, true) // role cookie ekledik

	// Rol kontrolü ve yönlendirme
	if user.Role == "admin" {
		c.Redirect(http.StatusSeeOther, "/admin")
	} else {
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// GET /register
func ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"Error":    "",
		"Success":  "",
		"Username": "",
		"Email":    "",
	})
}

// POST /register
func Register(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if username == "" || email == "" || password == "" {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"Error":    "Lütfen tüm alanları doldurun",
			"Username": username,
			"Email":    email,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"Error":    "Şifre oluşturulamadı",
			"Username": username,
			"Email":    email,
		})
		return
	}

	user := models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		IsPremium: false,
		Role:      "user",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"Error":    "Kullanıcı oluşturulamadı: " + err.Error(),
			"Username": username,
			"Email":    email,
		})
		return
	}

	c.HTML(http.StatusOK, "register.html", gin.H{
		"Success": "Kayıt başarılı! Giriş yapabilirsiniz.",
	})
}

// Dashboard (GET /)
func Dashboard(c *gin.Context) {
	userID, err := c.Cookie("user_id")
	if err != nil {
		// Çerez yoksa login sayfasına yönlendir
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Veritabanından kullanıcıyı getir
	uid, _ := strconv.Atoi(userID)
	var user models.User
	if err := config.DB.First(&user, uid).Error; err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Sayfayı render et, kullanıcı ve premium durumu gönder
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Username":  user.Username,
		"IsPremium": user.IsPremium,
	})
}

// Premium satın alma (POST /buy-premium)
func BuyPremium(c *gin.Context) {
	userID, err := c.Cookie("user_id")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	uid, _ := strconv.Atoi(userID)
	var user models.User
	if err := config.DB.First(&user, uid).Error; err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Demo amaçlı direkt premium yapıyoruz, normalde ödeme altyapısı kullanılır
	user.IsPremium = true
	if err := config.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{
			"Error": "Premium satın alınamadı.",
		})
		return
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Success":   "Premium başarıyla satın alındı!",
		"Username":  user.Username,
		"IsPremium": user.IsPremium,
	})
}

func Logout(c *gin.Context) {
	c.SetCookie("user_id", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/login")
}
