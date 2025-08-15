package routes

import (
	"fmt"
	"net/http"

	"mywetransfergo/config"
	"mywetransfergo/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

func SendFileEmail(c *gin.Context) {
	sender := c.PostForm("sender_email")
	receiver := c.PostForm("receiver_email")
	subject := c.PostForm("subject")

	// Yüklenen dosyayı DB'ye kaydet
	file, err := c.FormFile("file")
	if err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Dosya alınamadı"})
		return
	}

	filePath := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Dosya kaydedilemedi"})
		return
	}

	// DB kaydı
	newFile := models.File{
		Name: file.Filename,
		Path: filePath,
	}
	if err := config.DB.Create(&newFile).Error; err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Dosya DB'ye kaydedilemedi"})
		return
	}

	// İndirme linkini oluştur
	downloadLink := fmt.Sprintf("http://go.esrayaman.com.tr/download/%d", newFile.ID)
	body := fmt.Sprintf(`
Merhaba,

Dosyanız hazır. Aşağıdaki linke tıklayarak indirebilirsiniz:

<a href="%s">Dosyayı indir</a>

İyi günler.
`, downloadLink)

	// Mail gönderimi
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body) // HTML formatı

	d := gomail.NewDialer(
		"smtp.gmail.com",
		587,
		"esrayaman2332@gmail.com",
		"dgqrxxncyyhqasar",
	)

	if err := d.DialAndSend(m); err != nil {
		c.HTML(http.StatusOK, "base.html", gin.H{"Error": "Mail gönderilemedi: " + err.Error()})
		return
	}

	c.HTML(http.StatusOK, "base.html", gin.H{"Success": "Mail başarıyla gönderildi!"})
}
