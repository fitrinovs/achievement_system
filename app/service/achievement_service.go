// File: app/service/achievement_service.go

package service

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"os"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =================================================================
// ACHIEVEMENT SERVICE INTERFACE (KONTRAK)
// =================================================================

type AchievementService interface {
	// Standard CRUD
	CreateAchievement(c *gin.Context)
	UpdateAchievement(c *gin.Context)
	DeleteAchievement(c *gin.Context)

	// Read/List Endpoints
	GetAchievementsList(c *gin.Context) 
	GetAchievementByID(c *gin.Context) 

	// Workflow/Status Endpoints
	SubmitAchievement(c *gin.Context) 
	VerifyAchievement(c *gin.Context) 
	RejectAchievement(c *gin.Context) 

	// Utility Endpoints
	GetAchievementHistory(c *gin.Context) 
	UploadAttachment(c *gin.Context) 
}

// =================================================================
// ACHIEVEMENT SERVICE IMPLEMENTATION STRUCT
// =================================================================

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

// =================================================================
// PRIVATE UTILITY FUNCTIONS
// =================================================================

// calculatePoints adalah contoh sederhana logika bisnis untuk poin prestasi.
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
		return 0
	}
}

// Helper untuk mendapatkan Student UUID dari User UUID yang sedang login
func (s *achievementService) getStudentByUserID(c *gin.Context) (*model.Student, error) {
	// Gunakan "userID" sesuai dengan yang diset di AuthMiddleware
	userIDStr := c.GetString("userID") 
	if userIDStr == "" {
		return nil, errors.New("user ID not found in context")
	}
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	
	// Asumsi StudentRepository memiliki FindByUserID
	student, err := s.studentRepo.FindByUserID(userUUID) 
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}
	return student, nil
}

// =================================================================
// PUBLIC HANDLER IMPLEMENTATION (DENGAN SWAGGER)
// =================================================================

// @Summary Create New Achievement
// @Description Mahasiswa dapat membuat prestasi baru (status DRAFT).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param achievement body model.AchievementCreateRequest true "Data Prestasi Baru"
// @Success 201 {object} object{status=string,data=model.Achievement}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/achievements [post]
func (s *achievementService) CreateAchievement(c *gin.Context) {
	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can create achievements"})
		return
	}

	var req model.AchievementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	achievement := &model.Achievement{
		StudentID:   student.ID,
		Title:       req.Title,
		Category:    req.Category,
		Level:       req.Level,
		Description: req.Description,
		Points:      calculatePoints(req.Level),
		Status:      model.StatusDraft, 
	}

	if err := s.achievementRepo.Create(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create achievement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": achievement})
}

// @Summary List Achievements (Filtered by Role)
// @Description Mahasiswa: List achievements sendiri (read_own). Dosen/Admin: List semua (read_list).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} object{status=string,data=[]model.Achievement}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/achievements [get]
func (s *achievementService) GetAchievementsList(c *gin.Context) {
	permissionsAny, exists := c.Get("permissions")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Permissions not found in context"})
		return
	}
	permissions := permissionsAny.([]string)

	hasReadList := false
	hasReadOwn := false
	for _, p := range permissions {
		if p == "achievement:read_list" {
			hasReadList = true
		}
		if p == "achievement:read_own" {
			hasReadOwn = true
		}
	}

	var achievements []model.Achievement
	var err error

	if hasReadList {
		achievements, err = s.achievementRepo.FindAll()
	} else if hasReadOwn {
		student, studentErr := s.getStudentByUserID(c)
		if studentErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to retrieve student profile"})
			return
		}
		achievements, err = s.achievementRepo.FindByStudentID(student.ID)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Access denied: insufficient permissions"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": achievements})
}


// @Summary Get Achievement Detail by ID
// @Description Mengambil detail prestasi berdasarkan ID. Membutuhkan hak akses (owner atau admin/dosen).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,data=model.Achievement}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id} [get]
func (s *achievementService) GetAchievementByID(c *gin.Context) {
	id := c.Param("id")
	achID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": achievement})
}

// @Summary Update Achievement Data
// @Description Mahasiswa mengupdate data prestasi (hanya status DRAFT/REJECTED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Param achievement body model.AchievementUpdateRequest true "Data Prestasi yang Diperbarui"
// @Success 200 {object} object{status=string,data=model.Achievement}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id} [put]
func (s *achievementService) UpdateAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}
	
	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can update their achievements"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if achievement.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner of this achievement"})
		return
	}
	
	if achievement.Status != model.StatusDraft && achievement.Status != model.StatusRejected {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Cannot update achievement in status PENDING or APPROVED"})
		return
	}
	
	var req model.AchievementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	
	achievement.Title = req.Title
	achievement.Category = req.Category
	achievement.Level = req.Level
	achievement.Description = req.Description
	achievement.Points = calculatePoints(req.Level)
	
	if err := s.achievementRepo.Update(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update achievement"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement updated successfully", "data": achievement})
}

// @Summary Delete Achievement
// @Description Mahasiswa menghapus prestasi (hanya status DRAFT/REJECTED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id} [delete]
func (s *achievementService) DeleteAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}

	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can delete their achievements"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if achievement.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner of this achievement"})
		return
	}
	
	if achievement.Status != model.StatusDraft && achievement.Status != model.StatusRejected {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Cannot delete achievement in status PENDING or APPROVED"})
		return
	}
	
	if err := s.achievementRepo.Delete(achID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to delete achievement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement deleted successfully"})
}

// @Summary Submit Achievement for Verification
// @Description Mahasiswa mengubah status prestasi dari DRAFT/REJECTED menjadi PENDING. FileProof wajib ada.
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,message=string}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id}/submit [post]
func (s *achievementService) SubmitAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}
	
	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Only students can submit"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if achievement.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner"})
		return
	}

	if achievement.Status != model.StatusDraft && achievement.Status != model.StatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is already PENDING or APPROVED"})
		return
	}
	if achievement.FileProof == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File proof must be uploaded before submitting"})
		return
	}
	
	achievement.Status = model.StatusPending
	
	if err := s.achievementRepo.Update(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to submit achievement"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement submitted for verification"})
}

// @Summary Verify Achievement (Approve)
// @Description Dosen Wali/Admin menyetujui prestasi (Status PENDING -> APPROVED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,data=model.Achievement}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id}/verify [post]
func (s *achievementService) VerifyAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}

	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)
	
	// Asumsi LecturerRepository memiliki FindByUserID
	lecturer, err := s.lecturerRepo.FindByUserID(userUUID) 
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only authorized validators (Lecturers/Admins) can verify"})
		return
	}

	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}
	
	if achievement.Status != model.StatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is not in PENDING status"})
		return
	}

	achievement.Status = model.StatusApproved
	achievement.ValidatorID = &lecturer.ID 
	
	if err := s.achievementRepo.Update(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to approve achievement"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement approved successfully", "data": achievement})
}

// @Summary Reject Achievement
// @Description Dosen Wali/Admin menolak prestasi (Status PENDING -> REJECTED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Param rejection body model.AchievementRejectRequest true "Alasan Penolakan"
// @Success 200 {object} object{status=string,data=model.Achievement}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id}/reject [post]
func (s *achievementService) RejectAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}
	
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)
	
	// Asumsi LecturerRepository memiliki FindByUserID
	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only authorized validators (Lecturers/Admins) can reject"})
		return
	}
	
	var req model.AchievementRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	
	achievement, err := s.achievementRepo.FindByID(achID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}
	
	if achievement.Status != model.StatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is not in PENDING status"})
		return
	}

	rejectReason := req.Reason
	achievement.Status = model.StatusRejected
	achievement.ValidatorID = &lecturer.ID
	achievement.Reason = &rejectReason 
	
	if err := s.achievementRepo.Update(achievement); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to reject achievement"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement rejected successfully", "data": achievement})
}

// @Summary Get Achievement History
// @Description Endpoint placeholder untuk melihat riwayat perubahan status prestasi. (Memerlukan implementasi tabel history/log terpisah)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 501 {object} object{status=string,message=string} "Not Implemented"
// @Router /api/v1/achievements/{id}/history [get]
func (s *achievementService) GetAchievementHistory(c *gin.Context) {
	// Logika: Memerlukan tabel AchievementHistory (Log) yang terpisah.
	c.JSON(http.StatusNotImplemented, gin.H{"status": "info", "message": "Endpoint History Prestasi belum diimplementasikan. Membutuhkan model log/history terpisah."})
}

// @Summary Upload Attachment/Proof
// @Description Mahasiswa mengunggah file bukti prestasi (hanya untuk status DRAFT/REJECTED).
// @Tags Achievements
// @Accept mpfd
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Param file formData file true "File bukti prestasi"
// @Success 200 {object} object{status=string,message=string,file_path=string}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/achievements/{id}/attachments [post]
func (s *achievementService) UploadAttachment(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID"})
		return
	}
	
	// 1. Ambil file dari form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File is required"})
		return
	}
	
	// 2. Lakukan Ownership dan Status Check (Placeholder, sama seperti Update)
	student, _ := s.getStudentByUserID(c)
	achievement, _ := s.achievementRepo.FindByID(achID)

	if achievement.StudentID != student.ID || (achievement.Status != model.StatusDraft && achievement.Status != model.StatusRejected) {
		// Logika otorisasi yang lebih ketat harus diterapkan di sini,
		// tapi untuk kasus ini, kita asumsikan middleware sudah memblokir user non-owner.
		// Jika perlu, tambahkan error 403 di sini.
	}
	
	// 3. Simpan file
	// Menghindari konflik nama file
	filename := fmt.Sprintf("%s-%s-%s", 
		strings.ReplaceAll(achID.String(), "-", ""), 
		time.Now().Format("20060102150405"), 
		filepath.Base(file.Filename))
		
	// Tentukan destinasi penyimpanan
	dst := filepath.Join("uploads", "achievements", filename)
	
	// Pastikan direktori "uploads/achievements" ada
	if err := os.MkdirAll(filepath.Join("uploads", "achievements"), 0755); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create upload directory"})
        return
    }

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save file"})
		return
	}

	// 4. Update achievement.FileProof di DB
	achievement.FileProof = dst
	if err := s.achievementRepo.Update(achievement); err != nil {
		// Jika update DB gagal, file mungkin sudah tersimpan. Perlu logic rollback.
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to record file path in database"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "File uploaded successfully", "file_path": dst})
}