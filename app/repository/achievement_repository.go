package repository

import (
	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	Create(achievement *model.Achievement) error
	FindByID(id uuid.UUID) (*model.Achievement, error)
	FindAll() ([]model.Achievement, error)
	FindByStudentID(studentID uuid.UUID) ([]model.Achievement, error)
	Update(achievement *model.Achievement) error
	Delete(id uuid.UUID) error
}

type achievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) Create(achievement *model.Achievement) error {
	return r.db.Create(achievement).Error
}

func (r *achievementRepository) FindByID(id uuid.UUID) (*model.Achievement, error) {
	var achievement model.Achievement
	// Preload Student agar data mahasiswa juga ikut terambil
	err := r.db.Preload("Student.User").Preload("Validator.User").First(&achievement, "id = ?", id).Error
	return &achievement, err
}

func (r *achievementRepository) FindAll() ([]model.Achievement, error) {
	var achievements []model.Achievement
	err := r.db.Preload("Student.User").Find(&achievements).Error
	return achievements, err
}

func (r *achievementRepository) FindByStudentID(studentID uuid.UUID) ([]model.Achievement, error) {
	var achievements []model.Achievement
	err := r.db.Where("student_id = ?", studentID).Find(&achievements).Error
	return achievements, err
}

func (r *achievementRepository) Update(achievement *model.Achievement) error {
	return r.db.Save(achievement).Error
}

func (r *achievementRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Achievement{}, id).Error
}
