package repository

import (
	"errors"
	"fmt" // Import fmt untuk error yang lebih informatif

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =================================================================
// STUDENT REPOSITORY INTERFACE
// =================================================================

type StudentRepository interface {
	FindAll() ([]model.Student, error)
	FindByID(id uuid.UUID) (*model.Student, error)
	FindByUserID(userID uuid.UUID) (*model.Student, error)
	FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error)
	
	Create(student *model.Student) error
	Update(student *model.Student) error
	Delete(id uuid.UUID) error
	AssignAdvisor(studentID, advisorID uuid.UUID) error

	// METHOD BARU UNTUK REPORT SERVICE
	FindAdviseeIDsByAdvisorID(advisorID uuid.UUID) ([]uuid.UUID, error) 
	FindByStudentID(studentID string) (*model.Student, error) 
}

// =================================================================
// STUDENT REPOSITORY IMPLEMENTATION
// =================================================================

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepository{db}
}

func (r *studentRepository) FindAll() ([]model.Student, error) {
	var students []model.Student
	err := r.db.Preload("User").
		Preload("Advisor.User").
		Find(&students).Error
	return students, err
}

func (r *studentRepository) FindByID(id uuid.UUID) (*model.Student, error) {
	var student model.Student
	if err := r.db.Preload("User").
		Preload("Advisor.User").
		First(&student, "id = ?", id).Error; err != nil {
		return nil, errors.New("student not found")
	}
	return &student, nil
}

func (r *studentRepository) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	// Tambahkan Preload User agar bisa digunakan di report service
	if err := r.db.Preload("User").First(&student, "user_id = ?", userID).Error; err != nil {
		return nil, errors.New("student not found")
	}
	return &student, nil
}

func (r *studentRepository) FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Where("advisor_id = ?", advisorID).
		Preload("User").
		Find(&students).Error
	return students, err
}

func (r *studentRepository) Create(student *model.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepository) Update(student *model.Student) error {
	return r.db.Save(student).Error
}

func (r *studentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Student{}, id).Error
}

func (r *studentRepository) AssignAdvisor(studentID, advisorID uuid.UUID) error {
	return r.db.Model(&model.Student{}).
		Where("id = ?", studentID).
		Update("advisor_id", advisorID).Error
}


// =================================================================
// IMPLEMENTASI METHOD BARU UNTUK REPORT SERVICE
// =================================================================

// FindAdviseeIDsByAdvisorID mengambil daftar UUID mahasiswa yang dibimbing oleh dosen ini.
func (r *studentRepository) FindAdviseeIDsByAdvisorID(advisorID uuid.UUID) ([]uuid.UUID, error) {
	var studentIDs []uuid.UUID
	
	// Menggunakan Pluck untuk mengambil daftar kolom 'id' (UUID Student) saja
	err := r.db.
		Model(&model.Student{}).
		Where("advisor_id = ?", advisorID).
		Pluck("id", &studentIDs).
		Error
		
	// Tangani jika tidak ada record yang ditemukan
	if errors.Is(err, gorm.ErrRecordNotFound) {
        return []uuid.UUID{}, nil 
    }
	if err != nil {
		return nil, fmt.Errorf("failed to find advisee IDs: %w", err) 
	}
	return studentIDs, nil
}

// FindByStudentID mencari detail mahasiswa berdasarkan NIM/Student ID string.
func (r *studentRepository) FindByStudentID(studentID string) (*model.Student, error) {
	var student model.Student
	
	// Preload User diperlukan agar student.User.FullName bisa diakses di service
	if err := r.db.
		Preload("User"). // Pastikan relasi User dimuat
		Where("student_id = ?", studentID).
		First(&student).
		Error; err != nil {
		
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("student not found with the given student ID (NIM)")
        }
		return nil, fmt.Errorf("failed to find student by student ID: %w", err)
	}
	return &student, nil
}