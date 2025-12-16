// File: app/model/achievement.go

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =================================================================
// KONSTANTA STATUS PRESTASI
// =================================================================

type AchievementStatus string

const (
	StatusDraft    AchievementStatus = "DRAFT"    // Status awal, bisa diubah/hapus oleh Mahasiswa
	StatusPending  AchievementStatus = "PENDING"  // Status setelah di-submit, menunggu verifikasi
	StatusApproved AchievementStatus = "APPROVED" // Status disetujui oleh Dosen Wali/Validator
	StatusRejected AchievementStatus = "REJECTED" // Status ditolak, bisa di-update/re-submit oleh Mahasiswa
)

// =================================================================
// STRUCT UTAMA: ACHIEVEMENT
// =================================================================

type Achievement struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// Foreign Key ke Student (Pemilik Prestasi)
	StudentID uuid.UUID `gorm:"type:uuid;not null" json:"student_id"`
	// Asumsi model Student dan User sudah terdefinisikan.
	Student *Student `gorm:"foreignKey:StudentID;references:ID" json:"student,omitempty"` 

	Title       string `gorm:"not null" json:"title"`
	Category    string `gorm:"not null" json:"category"`
	Level       string `gorm:"not null" json:"level"`
	Points      int    `gorm:"default:0" json:"points"`
	Description string `gorm:"type:text" json:"description"`
	
	// File bukti, wajib diisi sebelum status PENDING
	FileProof string `json:"file_proof"` 

	// Status Workflow
	Status AchievementStatus `gorm:"type:varchar(20);default:'DRAFT'" json:"status"` // Default menjadi DRAFT
	
	// Validator (Dosen Wali/Admin)
	ValidatorID *uuid.UUID `gorm:"type:uuid" json:"validator_id"`
	// Asumsi model Lecturer sudah terdefinisikan.
	Validator   *Lecturer  `gorm:"foreignKey:ValidatorID" json:"validator,omitempty"` 
	
	// Alasan penolakan, hanya diisi jika Status = StatusRejected
	Reason *string `gorm:"type:text" json:"reason,omitempty"` 
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// =================================================================
// DTO (DATA TRANSFER OBJECTS)
// =================================================================

// AchievementCreateRequest: DTO untuk POST /api/v1/achievements
// Mahasiswa membuat Achievement awal (Status DRAFT)
type AchievementCreateRequest struct {
	Title       string `json:"title" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Level       string `json:"level" binding:"required"`
	Description string `json:"description"`
	// Note: FileProof di-upload terpisah (POST /attachments) atau ditangani sebagai form-data.
}

// AchievementUpdateRequest: DTO untuk PUT /api/v1/achievements/:id
// Mahasiswa mengedit data prestasi yang masih DRAFT atau REJECTED.
type AchievementUpdateRequest struct {
	Title       string `json:"title" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Level       string `json:"level" binding:"required"`
	Description string `json:"description"`
}

// AchievementRejectRequest: DTO untuk POST /api/v1/achievements/:id/reject
// Dosen Wali/Validator menolak Prestasi.
type AchievementRejectRequest struct {
	Reason string `json:"reason" binding:"required"` // Alasan penolakan wajib diisi
}

// Catatan: DTO AchievementValidateRequest lama (jika ada) sudah diganti
// oleh VerifyAchievement (tanpa body) dan RejectAchievement (membutuhkan Reason).