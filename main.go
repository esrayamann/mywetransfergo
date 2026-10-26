package main

import (
	"log"
	"mywetransfergo/config"
	"mywetransfergo/models"
	"mywetransfergo/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Uygulama yapılandırması geçersiz:", err)
	}
	if err := config.ConnectDB(); err != nil {
		log.Fatal(err)
	}

	// DB migrate
	err := config.DB.AutoMigrate(&models.User{}, &models.File{})
	if err != nil {
		log.Fatal("DB migrate hatası:", err)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), routes.SecurityHeaders())
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = config.App.MaxUploadBytes

	r.GET("/", routes.Home)
	r.GET("/health", routes.HealthCheck)
	r.GET("/download/:id", routes.DownloadFile)

	r.GET("/login", routes.ShowLoginPage)
	r.POST("/login", routes.Login)
	r.GET("/register", routes.ShowRegisterPage)
	r.POST("/register", routes.Register)

	protected := r.Group("/")
	protected.Use(routes.RequireUser)
	protected.GET("/dashboard", routes.Dashboard)
	protected.POST("/logout", routes.RequireCSRF, routes.Logout)
	protected.POST("/upload", routes.RequireCSRF, routes.UploadFile)
	protected.POST("/send-file", routes.RequireCSRF, routes.SendFileEmail)
	protected.POST("/buy-premium", routes.RequireCSRF, routes.BuyPremium)

	admin := r.Group("/")
	admin.Use(routes.RequireAdmin)
	admin.GET("/admin", routes.AdminPage)
	admin.POST("/change-role", routes.RequireCSRF, routes.ChangeUserRole)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("Sunucu dinleniyor:", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Sunucu başlatılamadı:", err)
	}
}
