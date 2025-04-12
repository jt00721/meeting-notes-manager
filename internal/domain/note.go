package domain

import (
	"time"

	"gorm.io/gorm"
)

type Note struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Content     string `gorm:"not null"`
	Category    string `gorm:"index"`
	MeetingDate time.Time
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type NoteFilter struct {
	Keyword  string
	Category string
	FromDate *time.Time
	ToDate   *time.Time
}
