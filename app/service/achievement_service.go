package service

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AchievementService interface {
	CreateAchievement(c *gin.Context)
	GetAchievementByID(c *gin.Context)
	GetMyAchievements(c *gin.Context)
	ValidateAchievement(c *gin.Context)
}

type achievementService struct {
	achievementRepo repository.AchievementRepository
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository
}

func NewAchievementService(
	achievementRepo repository.AchievementRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) AchievementService {
	return &achievementService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
		lecturerRepo:    lecturerRepo,
	}
}

func calculatePoints(level string) int {
	switch level {
	case "Internasional":
		return 100
	case "Nasional":
		return 50
	case "Provinsi":
		return 25
	case "Universitas":
		return 10
	default:
		return 5
	}
}

// CreateAchievement godoc
// @Summary      Upload Achievement
// @Description  Student uploads a new achievement with file proof
// @Tags         Achievements
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "File Proof (PDF/JPG/PNG)"
// @Param        title formData string true "Achievement Title"
// @Param        category formData string true "Category (e.g. Lomba, Seminar)"
// @Param        level formData string true "Level (Internasional, Nasional, Provinsi, Universitas)"
// @Param        description formData string false "Description"
// @Success      201 {object} object{status=string,data=model.Achievement}
// @Failure      400 {object} object{status=string,message=string}
// @Failure      403 {object} object{status=string,message=string}
// @Failure      500 {object} object{status=string,message=string}
// @Router       /api/v1/achievements [post]
func (s *achievementService) CreateAchievement(c *gin.Context) {
	// 1. Ambil User ID dari Token
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid user session"})
		return
	}

	// 2. Cari Data Student
	student, err := s.studentRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "only students can upload achievements"})
		return
	}

	// 3. Bind Form Data
	var req model.AchievementCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 4. Handle File Upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "file proof is required"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".pdf" && ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "only pdf, jpg, png allowed"})
		return
	}

	// Gunakan student.NIM
	filename := fmt.Sprintf("%s_%d%s", student.NIM, time.Now().Unix(), ext)
	savePath := filepath.Join("uploads", "achievements", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to save file"})
		return
	}

	// 5. Simpan ke Database
	achievement := &model.Achievement{
		StudentID:   student.ID,
		Title:       req.Title,
		Category:    req.Category,
		Level:       req.Level,
		Points:      calculatePoints(req.Level),
		Description: req.Description,
		FileProof:   savePath,
		Status:      model.StatusPending,
	}

	if err := s.achievementRepo.Create(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": achievement})
}

// GetAchievementByID godoc
// @Summary      Get Achievement Detail
// @Description  Get specific achievement by UUID
// @Tags         Achievements
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Achievement UUID"
// @Success      200 {object} object{status=string,data=model.Achievement}
// @Failure      400 {object} object{status=string,message=string}
// @Failure      404 {object} object{status=string,message=string}
// @Router       /api/v1/achievements/{id} [get]
func (s *achievementService) GetAchievementByID(c *gin.Context) {
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid id"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "achievement not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": achievement})
}

// GetMyAchievements godoc
// @Summary      Get My Achievements
// @Description  Get all achievements uploaded by logged-in student
// @Tags         Achievements
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} object{status=string,data=[]model.Achievement}
// @Failure      403 {object} object{status=string,message=string}
// @Failure      500 {object} object{status=string,message=string}
// @Router       /api/v1/achievements/my [get]
func (s *achievementService) GetMyAchievements(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	student, err := s.studentRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "student not found"})
		return
	}

	achievements, err := s.achievementRepo.FindByStudentID(student.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": achievements})
}

// ValidateAchievement godoc
// @Summary      Validate Achievement
// @Description  Lecturer approves or rejects an achievement
// @Tags         Achievements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Achievement UUID"
// @Param        request body model.AchievementValidateRequest true "Validation Data"
// @Success      200 {object} object{status=string,message=string}
// @Failure      400 {object} object{status=string,message=string}
// @Failure      403 {object} object{status=string,message=string}
// @Failure      404 {object} object{status=string,message=string}
// @Router       /api/v1/achievements/{id}/validate [put]
func (s *achievementService) ValidateAchievement(c *gin.Context) {
	id := c.Param("id")
	achID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid id"})
		return
	}

	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "only lecturers can validate"})
		return
	}

	var req model.AchievementValidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "achievement not found"})
		return
	}

	achievement.Status = req.Status
	achievement.ValidatorID = &lecturer.ID

	if req.Points > 0 {
		achievement.Points = req.Points
	} else if req.Status == model.StatusRejected {
		achievement.Points = 0
	}

	if err := s.achievementRepo.Update(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "achievement validated"})
}
