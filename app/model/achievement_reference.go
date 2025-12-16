package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =================================================================
// KONSTANTA STATUS PRESTASI (Untuk PGSQL)
// =================================================================

// Pastikan konstanta ini sesuai dengan ENUM achievement_status di PostgreSQL
type AchievementStatus string

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
)

// =================================================================
// STRUCT UNTUK POSTGRESQL (achievement_references)
// =================================================================

// AchievementReference merepresentasikan tabel achievement_references di PostgreSQL
type AchievementReference struct {
	// ID (uuid), PRIMARY KEY
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	// Foreign Key ke Student (Pemilik Prestasi)
	StudentID uuid.UUID `gorm:"type:uuid;not null;index" json:"student_id"`
	// Relasi ke Student (opsional, tergantung kebutuhan preload)
	Student   *Student  `gorm:"foreignKey:StudentID" json:"student,omitempty"` 
	
	// Kunci Penghubung ke MongoDB
	MongoAchievementID string `gorm:"type:varchar(24);not null;uniqueIndex" json:"mongo_achievement_id"` 
	
	// Status Workflow
	Status AchievementStatus `gorm:"type:varchar(20);not null;index;default:'draft'" json:"status"`
	
	// Waktu Pengajuan dan Verifikasi
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`

	// Validator (Dosen Wali/Admin)
	VerifiedBy     *uuid.UUID `gorm:"type:uuid" json:"verified_by,omitempty"`
	VerifiedByUser *User      `gorm:"foreignKey:VerifiedBy" json:"verified_by_user,omitempty"`
	
	// Alasan penolakan
	RejectionNote *string `gorm:"type:text" json:"rejection_note,omitempty"` 
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Menentukan nama tabel secara eksplisit
func (AchievementReference) TableName() string {
	return "achievement_references"
}

// =================================================================
// DTOs yang BERHUBUNGAN DENGAN WORKFLOW (PGSQL)
// =================================================================

// DTO untuk POST /api/v1/achievements/:id/reject
type AchievementRejectRequest struct {
	RejectionNote string `json:"rejection_note" binding:"required"` 
}



// Catatan: DTO untuk CREATE/UPDATE prestasi sekarang harus ada di file achievement.go
// karena berhubungan dengan data MongoDB.