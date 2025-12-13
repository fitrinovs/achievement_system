package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	// Pastikan field ini bernama NIM
	NIM string `gorm:"type:varchar(50);unique;not null" json:"nim"`

	ProgramStudy string `gorm:"type:varchar(100)" json:"program_study"`
	AcademicYear string `gorm:"type:varchar(9)" json:"academic_year"`

	AdvisorID *uuid.UUID `gorm:"type:uuid" json:"advisor_id"`
	Advisor   *Lecturer  `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ==========================================
// PERBAIKAN DI SINI:
// Ganti field 'StudentID' menjadi 'NIM'
// ==========================================
type StudentCreateRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	NIM          string `json:"nim" binding:"required"` // <--- GANTI INI
	ProgramStudy string `json:"program_study" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
	AdvisorID    string `json:"advisor_id"`
}

type StudentUpdateRequest struct {
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    string `json:"advisor_id"`
}
