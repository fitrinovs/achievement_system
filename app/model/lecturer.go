package model

import (
	"time"

	"github.com/google/uuid"
)

// Lecturer merepresentasikan tabel lecturers di SRS
type Lecturer struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	LecturerID string    `json:"lecturer_id" gorm:"type:varchar(20);unique;not null"` // NIP/NIDN
	Department string    `json:"department" gorm:"type:varchar(100)"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Lecturer) TableName() string {
	return "lecturers"
}

// Struct untuk input pembuatan dosen baru
type LecturerCreateRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	LecturerID string `json:"lecturer_id" binding:"required"` // NIP
	Department string `json:"department"`
}

// Struct untuk update data dosen
type LecturerUpdateRequest struct {
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}
