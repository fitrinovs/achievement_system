package repository

import (
	"database/sql"
	"errors"

	"github.com/fitrinovs/achievement_system/app/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository interface {
	FindAll() ([]model.Student, error)
	FindByID(id uuid.UUID) (*model.Student, error)
	FindByUserID(userID uuid.UUID) (*model.Student, error)
	FindByStudentID(studentID string) (*model.Student, error)
	FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error)
	Create(student *model.Student) error
	Update(student *model.Student) error
	Delete(id uuid.UUID) error
	AssignAdvisor(studentID, advisorID uuid.UUID) error
}

// ============ GORM IMPLEMENTATION ============

type studentRepositoryGORM struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepositoryGORM{db: db}
}

func (r *studentRepositoryGORM) FindAll() ([]model.Student, error) {
	var students []model.Student
	err := r.db.Preload("User").Preload("Advisor.User").Find(&students).Error
	return students, err
}

func (r *studentRepositoryGORM) FindByID(id uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor.User").Where("id = ?", id).First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

func (r *studentRepositoryGORM) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor.User").Where("user_id = ?", userID).First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

func (r *studentRepositoryGORM) FindByStudentID(studentID string) (*model.Student, error) {
	var student model.Student
	err := r.db.Preload("User").Preload("Advisor.User").Where("student_id = ?", studentID).First(&student).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

func (r *studentRepositoryGORM) FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error) {
	var students []model.Student
	err := r.db.Preload("User").Where("advisor_id = ?", advisorID).Find(&students).Error
	return students, err
}

func (r *studentRepositoryGORM) Create(student *model.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepositoryGORM) Update(student *model.Student) error {
	return r.db.Save(student).Error
}

func (r *studentRepositoryGORM) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Student{}, id).Error
}

func (r *studentRepositoryGORM) AssignAdvisor(studentID, advisorID uuid.UUID) error {
	return r.db.Model(&model.Student{}).
		Where("id = ?", studentID).
		Update("advisor_id", advisorID).Error
}

// ============ SQL IMPLEMENTATION ============

type StudentRepositorySQL struct {
	DB *sql.DB
}

func NewStudentRepositorySQL(db *sql.DB) StudentRepository {
	return &StudentRepositorySQL{DB: db}
}

func (r *StudentRepositorySQL) FindAll() ([]model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		ORDER BY student_id
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var student model.Student
		var advisorID sql.NullString

		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID,
			&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if advisorID.Valid {
			id, _ := uuid.Parse(advisorID.String)
			student.AdvisorID = &id
		}

		students = append(students, student)
	}

	return students, rows.Err()
}

func (r *StudentRepositorySQL) FindByID(id uuid.UUID) (*model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE id = $1
	`

	var student model.Student
	var advisorID sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&student.ID, &student.UserID, &student.StudentID,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if advisorID.Valid {
		advID, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &advID
	}

	return &student, nil
}

func (r *StudentRepositorySQL) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE user_id = $1
	`

	var student model.Student
	var advisorID sql.NullString

	err := r.DB.QueryRow(query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if advisorID.Valid {
		advID, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &advID
	}

	return &student, nil
}

func (r *StudentRepositorySQL) FindByStudentID(studentID string) (*model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE student_id = $1
	`

	var student model.Student
	var advisorID sql.NullString

	err := r.DB.QueryRow(query, studentID).Scan(
		&student.ID, &student.UserID, &student.StudentID,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	if advisorID.Valid {
		advID, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &advID
	}

	return &student, nil
}

func (r *StudentRepositorySQL) FindByAdvisorID(advisorID uuid.UUID) ([]model.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id = $1
		ORDER BY student_id
	`

	rows, err := r.DB.Query(query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var student model.Student
		var advID sql.NullString

		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID,
			&student.ProgramStudy, &student.AcademicYear, &advID, &student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if advID.Valid {
			id, _ := uuid.Parse(advID.String)
			student.AdvisorID = &id
		}

		students = append(students, student)
	}

	return students, rows.Err()
}

func (r *StudentRepositorySQL) Create(student *model.Student) error {
	query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	if student.ID == uuid.Nil {
		student.ID = uuid.New()
	}

	err := r.DB.QueryRow(
		query,
		student.ID, student.UserID, student.StudentID,
		student.ProgramStudy, student.AcademicYear, student.AdvisorID,
	).Scan(&student.CreatedAt)

	return err
}

func (r *StudentRepositorySQL) Update(student *model.Student) error {
	query := `
		UPDATE students 
		SET student_id = $1, program_study = $2, academic_year = $3, advisor_id = $4
		WHERE id = $5
	`

	result, err := r.DB.Exec(
		query,
		student.StudentID, student.ProgramStudy, student.AcademicYear,
		student.AdvisorID, student.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("student not found")
	}

	return nil
}

func (r *StudentRepositorySQL) Delete(id uuid.UUID) error {
	query := `DELETE FROM students WHERE id = $1`

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("student not found")
	}

	return nil
}

func (r *StudentRepositorySQL) AssignAdvisor(studentID, advisorID uuid.UUID) error {
	query := `UPDATE students SET advisor_id = $1 WHERE id = $2`

	result, err := r.DB.Exec(query, advisorID, studentID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("student not found")
	}

	return nil
}
