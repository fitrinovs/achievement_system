package service

import (
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
}

type lecturerService struct {
	lecturerRepo repository.LecturerRepository
	userRepo     repository.UserRepository
}

func NewLecturerService(lecturerRepo repository.LecturerRepository, userRepo repository.UserRepository) LecturerService {
	return &lecturerService{
		lecturerRepo: lecturerRepo,
		userRepo:     userRepo,
	}
}

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
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/lecturers [post]
func (s *lecturerService) CreateLecturer(c *gin.Context) {
	// 1. Parsing Request
	var req model.LecturerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 2. Validasi & Logic
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	// Cek apakah user ada
	_, err = s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	newLecturer := &model.Lecturer{
		UserID:     userUUID,
		LecturerID: req.LecturerID,
		Department: req.Department,
	}

	// 3. Simpan ke Database
	if err := s.lecturerRepo.Create(newLecturer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 4. Response JSON
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": newLecturer})
}

// GetLecturerByID godoc
// @Summary      Get Lecturer by ID
// @Tags         Lecturers
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Success      200  {object}  model.Lecturer
// @Router       /api/v1/lecturers/{id} [get]
func (s *lecturerService) GetLecturerByID(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id"})
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
// @Produce      json
// @Param        user_id   path      string  true  "User ID (UUID)"
// @Success      200  {object}  model.Lecturer
// @Router       /api/v1/lecturers/user/{user_id} [get]
func (s *lecturerService) GetLecturerByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id"})
		return
	}

	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": lecturer})
}

// GetAllLecturers godoc
// @Summary      Get All Lecturers
// @Tags         Lecturers
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   model.Lecturer
// @Router       /api/v1/lecturers [get]
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
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Param        request body model.LecturerUpdateRequest true "Data Update"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lecturers/{id} [put]
func (s *lecturerService) UpdateLecturer(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id"})
		return
	}

	var req model.LecturerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Cari data lama dulu
	lecturer, err := s.lecturerRepo.FindByID(lecturerUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "lecturer not found"})
		return
	}

	// Update field
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

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lecturer updated successfully"})
}

// DeleteLecturer godoc
// @Summary      Delete Lecturer
// @Tags         Lecturers
// @Security     BearerAuth
// @Param        id   path      string  true  "Lecturer ID (UUID)"
// @Success      200  {object}  map[string]string
// @Router       /api/v1/lecturers/{id} [delete]
func (s *lecturerService) DeleteLecturer(c *gin.Context) {
	id := c.Param("id")
	lecturerUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid lecturer id"})
		return
	}

	if err := s.lecturerRepo.Delete(lecturerUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Lecturer deleted successfully"})
}
