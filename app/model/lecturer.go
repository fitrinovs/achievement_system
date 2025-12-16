package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	User       *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	LecturerID string    `json:"lecturer_id" gorm:"type:varchar(20);unique;not null"`
	Department string    `json:"department" gorm:"type:varchar(100)"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (Lecturer) TableName() string {
	return "lecturers"
}

type LecturerCreateRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	LecturerID string `json:"lecturer_id" binding:"required"` // NIP
	Department string `json:"department"`
}

type LecturerUpdateRequest struct {
	LecturerID string `json:"lecturer_id"`
	Department string `json:"department"`
}

// ===================================
// Advisee (Model Response untuk GET /lecturers/:id/advisees)
// ===================================
type Advisee struct {
	ID           uuid.UUID `json:"id"`
	NIM          string    `json:"nim"`
	FullName     string    `json:"full_name"` // Diambil dari User yang di-preload
	ProgramStudy string    `json:"program_study"`
	AcademicYear string    `json:"academic_year"`
}