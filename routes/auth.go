package routes

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mywetransfergo/config"
	"mywetransfergo/models"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "mwt_session"
	sessionTTL        = 24 * time.Hour
	userContextKey    = "currentUser"
	claimsContextKey  = "sessionClaims"
)

type sessionClaims struct {
	UserID    uint   `json:"uid"`
	ExpiresAt int64  `json:"exp"`
	CSRFToken string `json:"csrf"`
}

func Home(c *gin.Context) {
	if _, err := CurrentUser(c); err == nil {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}

func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"Error": "", "Username": ""})
}

func Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	var user models.User
	if username == "" || password == "" || config.DB.Where("username = ? OR email = ?", username, username).First(&user).Error != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error":    "Kullanıcı adı/e-posta veya şifre hatalı.",
			"Username": username,
		})
		return
	}

	if err := startSession(c, user.ID); err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"Error": "Oturum başlatılamadı. Lütfen tekrar deneyin."})
		return
	}

	if user.Role == "admin" {
		c.Redirect(http.StatusSeeOther, "/admin")
		return
	}
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func ShowRegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"Error": "", "Success": "", "Username": "", "Email": "",
	})
}

func Register(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	email := strings.TrimSpace(c.PostForm("email"))
	password := c.PostForm("password")

	if len(username) < 3 || len(username) > 40 {
		renderRegisterError(c, "Kullanıcı adı 3-40 karakter arasında olmalı.", username, email)
		return
	}
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		renderRegisterError(c, "Geçerli bir e-posta adresi girin.", username, email)
		return
	}
	if len(password) < 10 {
		renderRegisterError(c, "Şifre en az 10 karakter olmalı.", username, email)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		renderRegisterError(c, "Şifre işlenemedi. Lütfen tekrar deneyin.", username, email)
		return
	}

	user := models.User{Username: username, Email: email, Password: string(hashedPassword), Role: "user"}
	if err := config.DB.Create(&user).Error; err != nil {
		renderRegisterError(c, "Bu kullanıcı adı veya e-posta zaten kullanılıyor.", username, email)
		return
	}

	c.HTML(http.StatusCreated, "register.html", gin.H{
		"Success":  "Kayıt başarılı. Şimdi giriş yapabilirsiniz.",
		"Username": "", "Email": "",
	})
}

func Dashboard(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	RenderDashboard(c, user, "", "")
}

func BuyPremium(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Ödeme entegrasyonu eklenene kadar bu işlem yalnızca demo amaçlıdır.
	user.IsPremium = true
	if err := config.DB.Save(user).Error; err != nil {
		RenderDashboard(c, user, "Premium durumu güncellenemedi.", "")
		return
	}
	RenderDashboard(c, user, "", "Premium durumu güncellendi.")
}

func Logout(c *gin.Context) {
	clearSession(c)
	c.Redirect(http.StatusSeeOther, "/login")
}

func RequireUser(c *gin.Context) {
	if _, err := CurrentUser(c); err != nil {
		clearSession(c)
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	c.Next()
}

func RequireAdmin(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		clearSession(c)
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	if user.Role != "admin" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}

func RequireCSRF(c *gin.Context) {
	claims, err := currentClaims(c)
	if err != nil || subtle.ConstantTimeCompare([]byte(c.PostForm("_csrf")), []byte(claims.CSRFToken)) != 1 {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	c.Next()
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; form-action 'self'; base-uri 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*models.User, error) {
	if value, exists := c.Get(userContextKey); exists {
		if user, ok := value.(*models.User); ok {
			return user, nil
		}
	}

	claims, err := currentClaims(c)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("oturum kullanıcısı bulunamadı")
	}
	c.Set(userContextKey, &user)
	return &user, nil
}

func RenderDashboard(c *gin.Context, user *models.User, errorMessage, successMessage string) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Username":  user.Username,
		"Email":     user.Email,
		"IsPremium": user.IsPremium,
		"CSRFToken": csrfToken(c),
		"Error":     errorMessage,
		"Success":   successMessage,
	})
}

func renderRegisterError(c *gin.Context, message, username, email string) {
	c.HTML(http.StatusUnprocessableEntity, "register.html", gin.H{
		"Error": message, "Username": username, "Email": email,
	})
}

func startSession(c *gin.Context, userID uint) error {
	csrfToken, err := randomToken(32)
	if err != nil {
		return err
	}
	claims := sessionClaims{UserID: userID, ExpiresAt: time.Now().Add(sessionTTL).Unix(), CSRFToken: csrfToken}
	token, err := signSession(claims)
	if err != nil {
		return err
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, token, int(sessionTTL.Seconds()), "/", "", config.App.CookieSecure, true)
	return nil
}

func clearSession(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", config.App.CookieSecure, true)
}

func currentClaims(c *gin.Context) (sessionClaims, error) {
	if value, exists := c.Get(claimsContextKey); exists {
		if claims, ok := value.(sessionClaims); ok {
			return claims, nil
		}
	}
	token, err := c.Cookie(sessionCookieName)
	if err != nil {
		return sessionClaims{}, errors.New("oturum çerezi bulunamadı")
	}
	claims, err := verifySession(token)
	if err != nil {
		return sessionClaims{}, err
	}
	c.Set(claimsContextKey, claims)
	return claims, nil
}

func csrfToken(c *gin.Context) string {
	claims, err := currentClaims(c)
	if err != nil {
		return ""
	}
	return claims.CSRFToken
}

func signSession(claims sessionClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, config.App.SessionSecret)
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func verifySession(token string) (sessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return sessionClaims{}, errors.New("geçersiz oturum biçimi")
	}

	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return sessionClaims{}, errors.New("geçersiz oturum imzası")
	}
	mac := hmac.New(sha256.New, config.App.SessionSecret)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(providedSignature, mac.Sum(nil)) {
		return sessionClaims{}, errors.New("oturum imzası doğrulanamadı")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return sessionClaims{}, errors.New("geçersiz oturum verisi")
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.UserID == 0 || claims.CSRFToken == "" || claims.ExpiresAt <= time.Now().Unix() {
		return sessionClaims{}, errors.New("oturum süresi dolmuş veya geçersiz")
	}
	return claims, nil
}

func randomToken(byteLength int) (string, error) {
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("rastgele değer üretilemedi: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
