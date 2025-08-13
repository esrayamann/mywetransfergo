package models

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	Password     string `gorm:"not null"`
	IsPremium    bool   `gorm:"default:false"`
	StorageQuota float64
	Role         string `gorm:"default:user"`
}
