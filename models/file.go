package models

import "time"

type File struct {
	ID        uint   `gorm:"primaryKey"`
	Filename  string `gorm:"not null"`
	Path      string `gorm:"not null"`
	Size      int64  `gorm:"not null"`
	Uploader  string // kullanıcı email ya da username
	CreatedAt time.Time
}
