# MyWeTransferGo

[![Go CI](https://github.com/esrayamann/mywetransfergo/actions/workflows/ci.yml/badge.svg)](https://github.com/esrayamann/mywetransfergo/actions/workflows/ci.yml)

Go ile geliştirilmiş, kayıtlı kullanıcıların dosya paylaşım bağlantısı oluşturup bunu e-posta ile gönderebildiği küçük bir dosya transfer uygulaması.

## Öne çıkanlar

- Gin ve GORM ile Go tabanlı web uygulaması
- PostgreSQL ile kullanıcı ve dosya meta verisi yönetimi
- Bcrypt ile parola karma işlemi
- İmzalı, `HttpOnly` oturum çerezi ve CSRF koruması
- Sunucu tarafında rol kontrolüyle korunan yönetici ekranı
- Boyut sınırı ve güvenli rastgele depolama adıyla yükleme akışı
- Tahmin edilmesi zor belirteç içeren indirme bağlantıları
- SMTP yapılandırması yalnızca ortam değişkenlerinden okunur

## Teknolojiler

Go 1.23 · Gin · GORM · PostgreSQL · SMTP

## Yerelde çalıştırma

Ön koşullar: Go 1.23+, PostgreSQL ve SMTP hesabı.

```bash
cp .env.example .env
# .env içindeki DATABASE_URL, SMTP_* ve SESSION_SECRET alanlarını kendi değerlerinle doldur.
go mod download
go run .
```

Uygulama varsayılan olarak `http://localhost:8081` adresinde çalışır. Başlatıldığında GORM gerekli şema güncellemelerini uygular.

## Yapılandırma

| Değişken | Açıklama |
| --- | --- |
| `DATABASE_URL` | PostgreSQL bağlantı adresi |
| `SMTP_HOST`, `SMTP_PORT` | SMTP sunucusu ve portu |
| `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` | E-posta gönderim kimlik bilgileri |
| `APP_BASE_URL` | İndirme bağlantılarında kullanılacak genel uygulama adresi |
| `SESSION_SECRET` | En az 32 karakterlik, her ortam için benzersiz imzalama anahtarı |
| `COOKIE_SECURE` | HTTPS arkasındaki üretim ortamında `true` |
| `MAX_UPLOAD_SIZE_MB` | Yüklenebilecek dosya sınırı; varsayılan `25` |

`.env.example` yalnızca şablondur. Gerçek `.env` dosyasını veya yüklenen dosyaları sürüm kontrolüne ekleme.

## Kontroller

```bash
go test ./...
go vet ./...
```

GitHub Actions bu kontrolleri `main` dalına gelen her push ve pull request'te çalıştırır.

## Güvenlik notları

- Uygulama e-posta gönderiminde formdaki gönderici adresini kullanmaz; yapılandırılmış SMTP hesabını kullanır.
- İndirme bağlantısı bir erişim belirteci taşır. Bağlantıyı yalnızca hedef kişiyle paylaş.
- Üretimde TLS sonlandırması kullan, `COOKIE_SECURE=true` ayarla ve dosyaları kalıcı disk ya da nesne depolama çözümünde sakla.
- Daha önce sürüm kontrolüne girmiş tüm e-posta veya veritabanı kimlik bilgilerini yenile.

## Lisans

[MIT](LICENSE)
