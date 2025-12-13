package service

import (
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StudentService interface {
	CreateStudent(c *gin.Context)
	GetStudentByID(c *gin.Context)
	GetStudentByUserID(c *gin.Context)
	GetAllStudents(c *gin.Context)
	GetStudentsByAdvisorID(c *gin.Context)
	UpdateStudent(c *gin.Context)
	DeleteStudent(c *gin.Context)
	AssignAdvisor(c *gin.Context)
}

type studentService struct {
	studentRepo  repository.StudentRepository
	userRepo     repository.UserRepository
	lecturerRepo repository.LecturerRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	lecturerRepo repository.LecturerRepository,
) StudentService {
	return &studentService{
		studentRepo:  studentRepo,
		userRepo:     userRepo,
		lecturerRepo: lecturerRepo,
	}
}

// CreateStudent godoc
// @Summary      Create Student
// @Tags         Students
// @Security     BearerAuth
// @Param        request body model.StudentCreateRequest true "Data Mahasiswa"
// @Router       /api/v1/students [post]
func (s *studentService) CreateStudent(c *gin.Context) {
	var req model.StudentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	// 1. Cek User Exist
	_, err = s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	// 2. Cek Duplicate Student ID (NIM)
	existing, _ := s.studentRepo.FindByStudentID(req.StudentID)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "student_id (NIM) already registered"})
		return
	}

	// 3. Cek Advisor jika ada
	var advisorUUID *uuid.UUID
	if req.AdvisorID != "" {
		parsedID, err := uuid.Parse(req.AdvisorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid advisor id format"})
			return
		}

		_, err = s.lecturerRepo.FindByID(parsedID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "advisor (lecturer) not found"})
			return
		}
		advisorUUID = &parsedID
	}

	student := &model.Student{
		UserID:       userUUID,
		StudentID:    req.StudentID,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
		AdvisorID:    advisorUUID,
	}

	if err := s.studentRepo.Create(student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": student})
}

// GetStudentByID godoc
// @Summary      Get Student by ID
// @Tags         Students
// @Security     BearerAuth
// @Param        id path string true "Student UUID"
// @Router       /api/v1/students/{id} [get]
func (s *studentService) GetStudentByID(c *gin.Context) {
	id := c.Param("id")
	studentUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	student, err := s.studentRepo.FindByID(studentUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": student})
}

// GetStudentByUserID godoc
// @Summary      Get Student by User ID
// @Tags         Students
// @Security     BearerAuth
// @Param        user_id path string true "User UUID"
// @Router       /api/v1/students/user/{user_id} [get]
func (s *studentService) GetStudentByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	student, err := s.studentRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": student})
}

// GetAllStudents godoc
// @Summary      Get All Students
// @Tags         Students
// @Security     BearerAuth
// @Router       /api/v1/students [get]
func (s *studentService) GetAllStudents(c *gin.Context) {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": students})
}

// GetStudentsByAdvisorID godoc
// @Summary      Get Students by Advisor
// @Tags         Students
// @Security     BearerAuth
// @Param        advisor_id path string true "Advisor UUID"
// @Router       /api/v1/students/advisor/{advisor_id} [get]
func (s *studentService) GetStudentsByAdvisorID(c *gin.Context) {
	advisorID := c.Param("advisor_id")
	advUUID, err := uuid.Parse(advisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid advisor id format"})
		return
	}
	students, err := s.studentRepo.FindByAdvisorID(advUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": students})
}

// UpdateStudent godoc
// @Summary      Update Student
// @Tags         Students
// @Security     BearerAuth
// @Param        id path string true "Student UUID"
// @Param        request body model.StudentUpdateRequest true "Data Update"
// @Router       /api/v1/students/{id} [put]
func (s *studentService) UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	studentUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	var req model.StudentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	student, err := s.studentRepo.FindByID(studentUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}

	if req.ProgramStudy != "" {
		student.ProgramStudy = req.ProgramStudy
	}
	if req.AcademicYear != "" {
		student.AcademicYear = req.AcademicYear
	}

	if req.AdvisorID != "" {
		advUUID, err := uuid.Parse(req.AdvisorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid advisor id format"})
			return
		}

		_, err = s.lecturerRepo.FindByID(advUUID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "advisor (lecturer) not found"})
			return
		}
		student.AdvisorID = &advUUID
	}

	if err := s.studentRepo.Update(student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Student updated successfully"})
}

// DeleteStudent godoc
// @Summary      Delete Student
// @Tags         Students
// @Security     BearerAuth
// @Param        id path string true "Student UUID"
// @Router       /api/v1/students/{id} [delete]
func (s *studentService) DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	studentUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	if err := s.studentRepo.Delete(studentUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Student deleted successfully"})
}

// AssignAdvisor godoc
// @Summary      Assign Advisor to Student
// @Tags         Students
// @Security     BearerAuth
// @Param        id path string true "Student UUID"
// @Param        request body object{advisor_id=string} true "Advisor UUID"
// @Router       /api/v1/students/{id}/advisor [put]
func (s *studentService) AssignAdvisor(c *gin.Context) {
	id := c.Param("id")
	sUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	// Ambil advisor_id dari JSON Body
	var req struct {
		AdvisorID string `json:"advisor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	aUUID, err := uuid.Parse(req.AdvisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid advisor id format"})
		return
	}

	// Cek exist advisor
	_, err = s.lecturerRepo.FindByID(aUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "advisor (lecturer) not found"})
		return
	}

	if err := s.studentRepo.AssignAdvisor(sUUID, aUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Advisor assigned successfully"})
}
