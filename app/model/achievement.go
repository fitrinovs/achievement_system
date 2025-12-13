package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementStatus string

const (
	StatusPending  AchievementStatus = "PENDING"
	StatusApproved AchievementStatus = "APPROVED"
	StatusRejected AchievementStatus = "REJECTED"
)

type Achievement struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// Foreign Key ke Student
	StudentID uuid.UUID `gorm:"type:uuid;not null" json:"student_id"`

	// =========================================================================
	// PERBAIKAN DI SINI:
	// Tambahkan ";references:ID" agar GORM tidak salah sambung ke NIM (string)
	// =========================================================================
	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"`

	Title    string `gorm:"not null" json:"title"`
	Category string `gorm:"not null" json:"category"`
	Level    string `gorm:"not null" json:"level"`
	Points   int    `gorm:"default:0" json:"points"`

	Description string `gorm:"type:text" json:"description"`
	FileProof   string `gorm:"not null" json:"file_proof"`

	Status      AchievementStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ValidatorID *uuid.UUID        `gorm:"type:uuid" json:"validator_id"`
	Validator   *Lecturer         `gorm:"foreignKey:ValidatorID" json:"validator,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ... (Struct DTO request di bawahnya biarkan saja, tidak perlu diubah)
type AchievementCreateRequest struct {
	Title       string `form:"title" binding:"required"`
	Category    string `form:"category" binding:"required"`
	Level       string `form:"level" binding:"required"`
	Description string `form:"description"`
}

type AchievementValidateRequest struct {
	Status      AchievementStatus `json:"status" binding:"required,oneof=APPROVED REJECTED"`
	Points      int               `json:"points"`
	ValidatorID string            `json:"-"`
}
