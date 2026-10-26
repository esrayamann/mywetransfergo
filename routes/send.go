package routes

import (
	"fmt"
	"mywetransfergo/utils"
	"net/http"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
)

func SendFileEmail(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	receiver := strings.TrimSpace(c.PostForm("receiver_email"))
	subject := strings.TrimSpace(c.PostForm("subject"))
	message := strings.TrimSpace(c.PostForm("body"))
	parsedAddress, err := mail.ParseAddress(receiver)
	if err != nil || parsedAddress.Address != receiver {
		RenderDashboard(c, user, "Geçerli bir alıcı e-posta adresi girin.", "")
		return
	}
	if subject == "" || len(subject) > 160 {
		RenderDashboard(c, user, "Konu alanı 1-160 karakter arasında olmalı.", "")
		return
	}
	if len(message) > 3000 {
		RenderDashboard(c, user, "Mesaj en fazla 3000 karakter olabilir.", "")
		return
	}

	file, err := storeUploadedFile(c)
	if err != nil {
		RenderDashboard(c, user, err.Error(), "")
		return
	}

	emailBody := fmt.Sprintf("Merhaba,\n\n%s\n\n%s tarafından paylaşılan %q dosyasını indirmek için aşağıdaki bağlantıyı kullanın:\n\n%s\n\nBu bağlantıyı yalnızca yetkili kişilerle paylaşın.", message, user.Email, file.Filename, downloadURL(file))
	if err := utils.SendMail(receiver, subject, emailBody); err != nil {
		RenderDashboard(c, user, "E-posta gönderilemedi. Lütfen SMTP ayarlarını kontrol edin.", "")
		return
	}

	RenderDashboard(c, user, "", "Dosya güvenli indirme bağlantısıyla e-posta olarak gönderildi.")
}
