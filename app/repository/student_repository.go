package repository

import (
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository interface {
	FindAll() ([]model.Student, error)
	FindByID(id uuid.UUID) (*model.Student, error)
	FindByUserID(userID uuid.UUID) (*model.Student, error)
	FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error)

	Create(student *model.Student) error
	Update(student *model.Student) error
	Delete(id uuid.UUID) error
	AssignAdvisor(studentID, advisorID uuid.UUID) error
}

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
	if err := r.db.First(&student, "user_id = ?", userID).Error; err != nil {
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
