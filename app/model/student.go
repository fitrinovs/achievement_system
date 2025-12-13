package model

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	User         *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	StudentID    string     `json:"student_id" gorm:"type:varchar(20);unique;not null"`
	ProgramStudy string     `json:"program_study" gorm:"type:varchar(100)"`
	AcademicYear string     `json:"academic_year" gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID `json:"advisor_id" gorm:"type:uuid"`
	Advisor      *Lecturer  `json:"advisor,omitempty" gorm:"foreignKey:AdvisorID"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:now()"`
}

func (Student) TableName() string {
	return "students"
}

// StudentCreateRequest digunakan di Service untuk membuat mahasiswa baru
type StudentCreateRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	StudentID    string `json:"student_id" binding:"required"` // NIM
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    string `json:"advisor_id"` // Optional
}

// StudentUpdateRequest digunakan di Service untuk update data mahasiswa
type StudentUpdateRequest struct {
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    string `json:"advisor_id"`
}
