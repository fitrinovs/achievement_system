package repository

import (
	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LecturerRepository interface {
	FindByID(id uuid.UUID) (*model.Lecturer, error)
	FindByUserID(userID uuid.UUID) (*model.Lecturer, error)
	FindAll() ([]*model.Lecturer, error)
	Create(lecturer *model.Lecturer) error
	Update(lecturer *model.Lecturer) error
	Delete(id uuid.UUID) error
}

type lecturerRepositoryGORM struct {
	db *gorm.DB
}

func NewLecturerRepository(db *gorm.DB) LecturerRepository {
	return &lecturerRepositoryGORM{db: db}
}

func (r *lecturerRepositoryGORM) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	// Preload User agar data profil dosen lengkap
	err := r.db.Preload("User").First(&lecturer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepositoryGORM) FindByUserID(userID uuid.UUID) (*model.Lecturer, error) {
	var lecturer model.Lecturer
	err := r.db.Preload("User").First(&lecturer, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &lecturer, nil
}

func (r *lecturerRepositoryGORM) FindAll() ([]*model.Lecturer, error) {
	var lecturers []*model.Lecturer
	err := r.db.Preload("User").Find(&lecturers).Error
	return lecturers, err
}

func (r *lecturerRepositoryGORM) Create(lecturer *model.Lecturer) error {
	return r.db.Create(lecturer).Error
}

func (r *lecturerRepositoryGORM) Update(lecturer *model.Lecturer) error {
	return r.db.Save(lecturer).Error
}

func (r *lecturerRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Lecturer{}, id).Error
}
