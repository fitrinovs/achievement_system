package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID" json:"user"`

	NIM           string `gorm:"type:varchar(20);unique;not null" json:"nim"`
	ProgramStudy  string `gorm:"type:varchar(100)" json:"program_study"`
	AcademicYear  string `gorm:"type:varchar(10)" json:"academic_year"`

	AdvisorID     *uuid.UUID `gorm:"type:uuid" json:"advisor_id"`
	Advisor       *Lecturer  `gorm:"foreignKey:AdvisorID" json:"advisor,omitempty"`

	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type StudentCreateRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	NIM          string `json:"nim" binding:"required"`
	ProgramStudy string `json:"program_study" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
}

type StudentUpdateRequest struct {
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
}
