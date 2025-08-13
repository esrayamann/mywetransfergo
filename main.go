package main

import (
	"log"
	"mywetransfergo/config"
	"mywetransfergo/models"
	"mywetransfergo/routes"
	"net/http"

	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ortam değişkenlerini yükle
	config.LoadEnv()
	config.ConnectDB()

	// DB migrate
	err := config.DB.AutoMigrate(&models.User{}, &models.File{})
	if err != nil {
		log.Fatal("DB migrate hatası:", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Endpointler
	r.GET("/health", routes.HealthCheck)
	r.POST("/upload", routes.UploadFile)
	r.GET("/download/:id", routes.DownloadFile)
	r.POST("/send-email", routes.SendFileEmail)
	r.POST("/send-file", routes.SendFileEmail)

	// Login & Register
	r.GET("/login", routes.ShowLoginPage)       // Login sayfası göster
	r.POST("/login", routes.Login)              // Login işlemi
	r.GET("/register", routes.ShowRegisterPage) // Register sayfası göster
	r.POST("/register", routes.Register)        // Register işlemi
	r.GET("/logout", routes.Logout)
	r.GET("/dashboard", routes.Dashboard)
	r.POST("/buy-premium", routes.BuyPremium)
	r.GET("/base", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base.html", gin.H{
			"Error":   "",
			"Success": "",
		})
	})
	r.GET("/admin", routes.AdminPage)
	r.POST("/change-role", routes.ChangeUserRole)

	// Port ayarı (nginx için 3000 yapıyoruz)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("Server çalışıyor port:", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Sunucu başlatılamadı:", err)
	}
}
