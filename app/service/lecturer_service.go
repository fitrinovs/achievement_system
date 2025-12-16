package service

import (
	"errors"
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LecturerService interface {
	CreateLecturer(c *gin.Context)
	GetLecturerByID(c *gin.Context)
	GetLecturerByUserID(c *gin.Context)
	GetAllLecturers(c *gin.Context)
	UpdateLecturer(c *gin.Context)
	DeleteLecturer(c *gin.Context)
	// SRS Endpoint:
	GetAdviseesByLecturerID(c *gin.Context)
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	userRepo     repository.UserRepository
	studentRepo  repository.StudentRepository
}

func NewLecturerService(
	lecturerRepo repository.LecturerRepository,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
		userRepo:     userRepo,
		studentRepo:  studentRepo,
	}
}

// ===================================
// Fix 2: Mengimplementasikan semua method interface LecturerService
// ===================================

// CreateLecturer godoc
// @Summary      Create Lecturer
// @Description  Menambahkan data dosen baru
// @Tags         Lecturers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body model.LecturerCreateRequest true "Data Dosen"
// @Success      201  {object}  model.Lecturer
// @Failure      400  {object}  map[string]string
// @Router       /lecturers [post]
func (s *lecturerService) CreateLecturer(c *gin.Context) {
	var req model.LecturerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid user id format"})
		return
	}

	// Cek user existence
	_, err = s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "User not found"})
		return
	}

	// Cek apakah lecturer sudah ada
	_, err = s.lecturerRepo.FindByUserID(userUUID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Lecturer already exists for this user"})
		return
	}

	newLecturer := model.Lecturer{
		UserID:     userUUID,
		LecturerID: req.LecturerID, // NIP/NIDN
		Department: req.Department,
	}

	if err := s.lecturerRepo.Create(&newLecturer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	lecturer, _ := s.lecturerRepo.FindByID(newLecturer.ID)

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": lecturer})
}

// GetLecturerByID godoc
// @Summary      Get Lecturer by ID
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id path string true "Lecturer UUID"
// @Success      200 {object} model.Lecturer
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /lecturers/{id} [get]
func (s *lecturerService) GetLecturerByID(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id format"})
		return
	}

	lecturer, err := s.lecturerRepo.FindByID(lecturerUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lecturer})
}

// GetLecturerByUserID godoc
// @Summary      Get Lecturer by User ID
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Success      200 {object} model.Lecturer
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /lecturers/user/{id} [get]
func (s *lecturerService) GetLecturerByUserID(c *gin.Context) {
	userID := c.Param("id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found for this user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lecturer})
}

// GetAllLecturers godoc
// @Summary      Get All Lecturers
// @Tags         Lecturers
// @Security     BearerAuth
// @Success      200 {array} model.Lecturer
// @Failure      500 {object} map[string]string
// @Router       /lecturers [get]
func (s *lecturerService) GetAllLecturers(c *gin.Context) {
	lecturers, err := s.lecturerRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lecturers})
}

// UpdateLecturer godoc
// @Summary      Update Lecturer
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id path string true "Lecturer UUID"
// @Param        request body model.LecturerUpdateRequest true "Lecturer Data"
// @Success      200 {object} model.Lecturer
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /lecturers/{id} [put]
func (s *lecturerService) UpdateLecturer(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id format"})
		return
	}

	var req model.LecturerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	lecturer, err := s.lecturerRepo.FindByID(lecturerUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
		return
	}

	if req.LecturerID != "" {
		lecturer.LecturerID = req.LecturerID
	}
	if req.Department != "" {
		lecturer.Department = req.Department
	}

	if err := s.lecturerRepo.Update(lecturer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lecturer updated successfully", "data": lecturer})
}

// DeleteLecturer godoc
// @Summary      Delete Lecturer
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router      /lecturers/{id} [delete]
func (s *lecturerService) DeleteLecturer(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id format"})
		return
	}

	if err := s.lecturerRepo.Delete(lecturerUUID); err != nil {
		if errors.Is(err, errors.New("lecturer not found")) {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lecturer deleted successfully"})
}

// ===================================
// IMPLEMENTASI SRS: GetAdviseesByLecturerID
// GET /lecturers/:id/advisees
// ===================================
// @Summary      Get Advisees (Students) by Lecturer ID
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id path string true "Lecturer UUID"
// @Success      200 {array} model.Advisee
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /lecturers/{id}/advisees [get]
func (s *lecturerService) GetAdviseesByLecturerID(c *gin.Context) {
	lecturerID := c.Param("id")
	lecturerUUID, err := uuid.Parse(lecturerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id format"})
		return
	}

	// 1. Cek apakah lecturer ada
	_, err = s.lecturerRepo.FindByID(lecturerUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
		return
	}

	// 2. Ambil students berdasarkan advisor_id
	students, err := s.studentRepo.FindByAdvisorID(lecturerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 3. Konversi ke format Advisee (model.Advisee)
	var advisees []model.Advisee
	for _, student := range students {
		fullname := ""
		// FIX 3: student.User adalah struct value, tidak bisa dibandingkan dengan nil. 
		// Kita cek ID-nya terhadap zero UUID (uuid.Nil).
		if student.User.ID != uuid.Nil {
			fullname = student.User.FullName 
		}

		advisees = append(advisees, model.Advisee{
			ID:           student.ID,
			NIM:          student.NIM,
			FullName:     fullname,
			ProgramStudy: student.ProgramStudy,
			AcademicYear: student.AcademicYear,
		})
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": advisees})
}