package service

import (
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =================================================================
// INTERFACE
// =================================================================

type StudentService interface {
	CreateStudent(c *gin.Context)
	GetAllStudents(c *gin.Context)
	GetStudentByID(c *gin.Context)
	UpdateStudent(c *gin.Context)
	DeleteStudent(c *gin.Context)

	AssignAdvisor(c *gin.Context)
	GetAchievementsByStudentID(c *gin.Context)
}

// =================================================================
// STRUCT IMPLEMENTATION
// =================================================================

type studentService struct {
	studentRepo     repository.StudentRepository
	userRepo        repository.UserRepository
	lecturerRepo    repository.LecturerRepository
	achievementRepo repository.AchievementRepository // Mengambil AchievementRepository
}

func NewStudentService(
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	lecturerRepo repository.LecturerRepository,
	achievementRepo repository.AchievementRepository,
) StudentService {
	return &studentService{
		studentRepo:     studentRepo,
		userRepo:        userRepo,
		lecturerRepo:    lecturerRepo,
		achievementRepo: achievementRepo,
	}
}

// =================================================================
// HANDLERS
// =================================================================

//
// =======================
// CREATE STUDENT
// =======================
// @Summary Create Student
// @Tags Students
// @Security BearerAuth
// @Param request body model.StudentCreateRequest true "Student Data"
// @Success 201 {object} model.Student
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/students [post]
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

	// Cek user
	if _, err := s.userRepo.FindByID(userUUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	// Cek apakah student sudah ada
	if _, err := s.studentRepo.FindByUserID(userUUID); err == nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "student already exists"})
		return
	}

	student := model.Student{
		UserID:       userUUID,
		NIM:          req.NIM,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
	}

	if err := s.studentRepo.Create(&student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	result, _ := s.studentRepo.FindByID(student.ID)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": result})
}

//
// =======================
// GET ALL STUDENTS
// =======================
// @Summary Get All Students
// @Tags Students
// @Security BearerAuth
// @Success 200 {array} model.Student
// @Router /api/v1/students [get]
func (s *studentService) GetAllStudents(c *gin.Context) {
	students, err := s.studentRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": students})
}

//
// =======================
// GET STUDENT BY ID
// =======================
// @Summary Get Student by ID
// @Tags Students
// @Security BearerAuth
// @Param id path string true "Student UUID"
// @Success 200 {object} model.Student
// @Failure 404 {object} map[string]string
// @Router /api/v1/students/{id} [get]
func (s *studentService) GetStudentByID(c *gin.Context) {
	studentUUID, err := uuid.Parse(c.Param("id"))
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

//
// =======================
// UPDATE STUDENT
// =======================
// @Summary Update Student
// @Tags Students
// @Security BearerAuth
// @Param id path string true "Student UUID"
// @Param request body model.StudentUpdateRequest true "Student Data"
// @Success 200 {object} model.Student
// @Router /api/v1/students/{id} [put]
func (s *studentService) UpdateStudent(c *gin.Context) {
	studentUUID, err := uuid.Parse(c.Param("id"))
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

	if err := s.studentRepo.Update(student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": student})
}

//
// =======================
// DELETE STUDENT
// =======================
// @Summary Delete Student
// @Tags Students
// @Security BearerAuth
// @Param id path string true "Student UUID"
// @Success 200 {object} map[string]string
// @Router /api/v1/students/{id} [delete]
func (s *studentService) DeleteStudent(c *gin.Context) {
	studentUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	if err := s.studentRepo.Delete(studentUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "student deleted"})
}

//
// =======================
// ASSIGN ADVISOR
// =======================
// @Summary Assign Advisor
// @Tags Students
// @Security BearerAuth
// @Param id path string true "Student UUID"
// @Param request body object{advisor_id=string} true "Advisor UUID"
// @Success 200 {object} map[string]string
// @Router /api/v1/students/{id}/advisor [put]
func (s *studentService) AssignAdvisor(c *gin.Context) {
	studentUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	var req struct {
		AdvisorID string `json:"advisor_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	advisorUUID, err := uuid.Parse(req.AdvisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid advisor id format"})
		return
	}

	if _, err := s.lecturerRepo.FindByID(advisorUUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "advisor not found"})
		return
	}

	if _, err := s.studentRepo.FindByID(studentUUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}

	if err := s.studentRepo.AssignAdvisor(studentUUID, advisorUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "advisor assigned"})
}

//
// =======================
// GET STUDENT ACHIEVEMENTS
// =======================
// @Summary Get Achievements by Student
// @Description Mengambil daftar referensi prestasi (PGSQL) milik mahasiswa.
// @Tags Students
// @Security BearerAuth
// @Param id path string true "Student UUID"
// @Success 200 {array} model.AchievementReference
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/students/{id}/achievements [get]
func (s *studentService) GetAchievementsByStudentID(c *gin.Context) {
	studentUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	if _, err := s.studentRepo.FindByID(studentUUID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}

	// PERBAIKAN UTAMA: Mengganti FindByStudentID yang tidak ada
	// dengan FindReferencesByStudentID yang kita definisikan di repository.
	// Outputnya adalah []model.AchievementReference (data PGSQL).
	achievements, err := s.achievementRepo.FindReferencesByStudentID(studentUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": achievements})
}