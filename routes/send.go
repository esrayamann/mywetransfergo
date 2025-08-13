package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

func SendFileEmail(c *gin.Context) {
	sender := c.PostForm("sender_email")
	receiver := c.PostForm("receiver_email")
	subject := c.PostForm("subject")
	body := c.PostForm("body")

	file, err := c.FormFile("file")
	if err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Dosya alınamadı"})
		return
	}

	// Dosyayı geçici olarak kaydet
	filePath := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Dosya kaydedilemedi"})
		return
	}

	// Mail gönderimi
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	m.Attach(filePath)

	d := gomail.NewDialer("smtp.gmail.com", 587, "esrayaman2332@gmail.com", "dgqrxxncyyhqasar")

	if err := d.DialAndSend(m); err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Mail gönderilemedi: " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "base.html", gin.H{"Success": "Mail başarıyla gönderildi!"})
}
