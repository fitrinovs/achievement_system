package model

import (
	"time"
	"github.com/google/uuid"
)

// =================================================================
// DTO UNTUK OUTPUT STATISTICS
// =================================================================

// StatisticsByGroup merepresentasikan hitungan berdasarkan kategori (tipe, level, periode)
type StatisticsByGroup struct {
	Group string `json:"group"` // Contoh: 'competition', 'Internasional', '2025'
	Count int    `json:"count"`
}

// TopStudentDetail merepresentasikan detail Top Mahasiswa Berprestasi
// Catatan: Ini digunakan oleh ReportRepository saat query ke PostgreSQL
type TopStudentDetail struct {
	// FIX: Menggunakan nim untuk konsistensi dengan model Student
	NIM           string `json:"nim"` 
	FullName      string `json:"full_name"`
	TotalPoints   int    `json:"total_points"`
}

// AchievementStatisticsResponse: Struktur utama untuk endpoint /api/v1/reports/statistics
type AchievementStatisticsResponse struct {
	TotalAchievementCount int                    `json:"total_achievement_count"`
	ByType                []StatisticsByGroup    `json:"by_type"`      // Total prestasi per tipe
	ByLevel               []StatisticsByGroup    `json:"by_level"`     // Distribusi tingkat kompetisi
	ByPeriod              []StatisticsByGroup    `json:"by_period"`    // Total prestasi per periode (tahun event)
	TopStudents           []TopStudentDetail     `json:"top_students"` // Top mahasiswa berprestasi
}

// StudentProfileResponse: Detail profil mahasiswa untuk laporan
type StudentProfileResponse struct {
	ID           uuid.UUID `json:"id"`
	NIM          string    `json:"nim"`           // FIX: Ganti dari StudentID ke NIM
	FullName     string    `json:"full_name"`
	ProgramStudy string    `json:"program_study"` // Field dari model.Student Anda
	AcademicYear string    `json:"academic_year"` // Field dari model.Student Anda
}

// StudentAchievementReportResponse: Struktur untuk endpoint /api/v1/reports/student/:id
type StudentAchievementReportResponse struct {
	StudentProfile *StudentProfileResponse `json:"student_profile"` 
	TotalPoints    int                     `json:"total_points"`
	// Asumsi model.Achievement sudah didefinisikan (digunakan untuk daftar prestasi)
	Achievements   []Achievement           `json:"achievements"` 
}

// =================================================================
// STRUKTUR INTERNAL (UNTUK QUERY DB DARI MONGODB)
// =================================================================

// MongoAchievementMinimal digunakan untuk menampung data minimal dari MongoDB
// yang diperlukan untuk perhitungan statistik (agregasi).
type MongoAchievementMinimal struct {
	ID              string    `bson:"_id"` 
	StudentUUID     uuid.UUID `bson:"studentId"`
	AchievementType string    `bson:"achievementType"`
	// Asumsi Level disimpan dalam details.competitionLevel di MongoDB
	Level           string    `bson:"details.competitionLevel"` 
	EventDate       time.Time `bson:"details.eventDate"`        
	Points          int       `bson:"points"` // Poin dari MongoDB (atau disinkronkan)
}