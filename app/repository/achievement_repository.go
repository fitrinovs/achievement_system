// File: app/repository/achievement_repository.go

package repository

import (
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =================================================================
// ACHIEVEMENT REPOSITORY INTERFACE (KONTRAK)
// =================================================================

// AchievementRepository mendefinisikan kontrak data access untuk entitas Achievement.
type AchievementRepository interface {
	Create(achievement *model.Achievement) error
	
	// Preload Student.User dan Validator.User agar data terkait ikut terambil
	FindByID(id uuid.UUID) (*model.Achievement, error) 
	
	FindAll() ([]model.Achievement, error)
	FindByStudentID(studentID uuid.UUID) ([]model.Achievement, error)

	// Digunakan untuk Update, Submit, Verify, Reject (Karena hanya update field status/data)
	Update(achievement *model.Achievement) error 
	
	Delete(id uuid.UUID) error
}

// =================================================================
// ACHIEVEMENT REPOSITORY IMPLEMENTATION
// =================================================================

type achievementRepository struct {
	db *gorm.DB
}

// NewAchievementRepository membuat instance baru dari AchievementRepository
func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &achievementRepository{db: db}
}

// =================================================================
// IMPLEMENTASI METHOD CRUD
// =================================================================

// Create menambahkan data prestasi baru ke database.
func (r *achievementRepository) Create(achievement *model.Achievement) error {
	return r.db.Create(achievement).Error
}

// FindByID mencari satu data prestasi berdasarkan ID-nya.
func (r *achievementRepository) FindByID(id uuid.UUID) (*model.Achievement, error) {
	var achievement model.Achievement
	
	// Preload Student.User dan Validator.User agar data terkait ikut terambil
	err := r.db.
		Preload("Student.User").
		Preload("Validator.User").
		First(&achievement, "id = ?", id).
		Error
		
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("achievement not found")
	}
	
	return &achievement, err
}

// Update menyimpan perubahan pada data prestasi yang sudah ada.
func (r *achievementRepository) Update(achievement *model.Achievement) error {
	// Pastikan hanya field yang dimodifikasi yang di-update (GORM Smart Select)
	// atau gunakan Save jika semua field ingin di-update
	return r.db.Save(achievement).Error
}

// Delete menghapus data prestasi secara soft delete (mengisi DeletedAt).
func (r *achievementRepository) Delete(id uuid.UUID) error {
	// Gunakan model.Achievement di sini untuk memicu soft delete
	return r.db.Delete(&model.Achievement{}, id).Error
}

// =================================================================
// IMPLEMENTASI METHOD READ KHUSUS
// =================================================================

// FindAll mengembalikan semua data prestasi (Biasanya untuk Admin/Dosen)
func (r *achievementRepository) FindAll() ([]model.Achievement, error) {
	var achievements []model.Achievement
	
	// Preload untuk list agar tidak terlalu berat, tapi tetap menampilkan user info
	err := r.db.
		Preload("Student.User").
		Preload("Validator.User").
		Find(&achievements).
		Error
		
	return achievements, err
}

// FindByStudentID mengembalikan daftar prestasi yang dimiliki oleh satu Mahasiswa.
func (r *achievementRepository) FindByStudentID(studentID uuid.UUID) ([]model.Achievement, error) {
	var achievements []model.Achievement
	
	// Preload yang dibutuhkan
	err := r.db.
		Preload("Student.User").
		Preload("Validator.User").
		Where("student_id = ?", studentID).
		Find(&achievements).
		Error
		
	return achievements, err
}

// Catatan: Jika Anda perlu fitur filtering khusus untuk Dosen Wali (misal: hanya prestasi mahasiswa bimbingannya), 
// Anda mungkin perlu menambahkan method baru seperti FindByLecturerID(lecturerID uuid.UUID)