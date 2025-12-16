// File: app/service/achievement_service.go (FINAL LENGKAP & REVISI ULTIMATE)

package service

import (
	"context"
	"encoding/json" // Diperlukan untuk konversi struct Details ke map
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		return nil, errors.New("user ID not found in context (Auth Middleware failed)")
	}
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	student, err := s.studentRepo.FindByUserID(userUUID)
	if err != nil {
		return nil, fmt.Errorf("student data not found for this user: %w", err)
	}
	if student == nil {
		return nil, errors.New("student data not found")
	}
	return student, nil
}

// Helper untuk menggabungkan data PGSQL dan MongoDB
func (s *achievementService) mergeAchievement(ctx context.Context, pgRef *model.AchievementReference) (*model.AchievementDetailResponse, error) {
	mongoAch, err := s.achievementRepo.FindMongoByID(ctx, pgRef.MongoAchievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to find achievement content in MongoDB: %w", err)
	}
	if mongoAch == nil {
		return nil, errors.New("data integrity error: mongo content missing")
	}

	return &model.AchievementDetailResponse{
		// PGSQL (Workflow)
		ID:            pgRef.ID,
		StudentID:     pgRef.StudentID,
		Status:        pgRef.Status,
		SubmittedAt:   pgRef.SubmittedAt,
		VerifiedAt:    pgRef.VerifiedAt,
		VerifiedBy:    pgRef.VerifiedBy,
		RejectionNote: pgRef.RejectionNote,

		// MongoDB (Content)
		MongoAchievementID: pgRef.MongoAchievementID,
		AchievementType:    mongoAch.AchievementType,
		Title:              mongoAch.Title,
		Description:        mongoAch.Description,
		Details:            mongoAch.Details,
		Attachments:        mongoAch.Attachments,
		Tags:               mongoAch.Tags,
		Points:             mongoAch.Points,

		CreatedAt: pgRef.CreatedAt,
		UpdatedAt: pgRef.UpdatedAt,
	}, nil
}

// =================================================================
// PUBLIC HANDLER IMPLEMENTATION
// =================================================================

// @Summary Create New Achievement
// @Description Mahasiswa dapat membuat prestasi baru (status DRAFT).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param achievement body model.AchievementCreateRequest true "Data Prestasi Baru"
// @Success 201 {object} object{status=string,data=model.AchievementDetailResponse}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /achievements [post]
func (s *achievementService) CreateAchievement(c *gin.Context) {
	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can create achievements: " + err.Error()})
		return
	}

	var req model.AchievementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body: " + err.Error()})
		return
	}

	// 1A. KONVERSI req.Details (STRUCT) ke map[string]interface{} (UNTUK MONGO)
	var detailsMap map[string]interface{}
	detailsBytes, err := json.Marshal(req.Details)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to marshal achievement details: " + err.Error()})
		return
	}

	if err := json.Unmarshal(detailsBytes, &detailsMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to unmarshal achievement details to map: " + err.Error()})
		return
	}

	// 1B. SIAPKAN DOKUMEN MONGO (Data Konten)
	mongoAchievement := &model.Achievement{
		StudentUUID:     student.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         detailsMap, // Menggunakan hasil konversi
		Points:          req.Points,
		Tags:            req.Tags,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Attachments:     []model.AttachmentMetadata{},
	}

	// 2. INSERT KE MONGODB
	mongoObjectID, err := s.achievementRepo.InsertMongoAchievement(c.Request.Context(), mongoAchievement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create achievement content (Mongo): " + err.Error()})
		return
	}

	// 3. SIAPKAN REFERENSI PGSQL (Workflow)
	pgReference := &model.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: mongoObjectID.Hex(),
		Status:             model.StatusDraft,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 4. INSERT KE POSTGRESQL
	if err := s.achievementRepo.CreateReference(pgReference); err != nil {
		// Rollback MongoDB jika PGSQL gagal (Best effort)
		s.achievementRepo.DeleteMongoAchievement(c.Request.Context(), pgReference.MongoAchievementID)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create achievement reference (PGSQL). Content rolled back."})
		return
	}

	// Response gabungan
	mergedAch, _ := s.mergeAchievement(c.Request.Context(), pgReference)
	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": mergedAch})
}

// @Summary List Achievements (Filtered by Role)
// @Description Mahasiswa: List achievements sendiri (read_own). Dosen/Admin: List semua (read_list). (Saat ini mengembalikan 501)
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} object{status=string,data=[]model.AchievementDetailResponse}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /achievements [get]
func (s *achievementService) GetAchievementsList(c *gin.Context) {
	// Implementasi List Hibrid (kompleks, perlu iterasi dan merge)

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

	var err error

	if hasReadList {
		// Implementasi s.achievementRepo.FindAllReferences() diperlukan
		err = errors.New("FindAllReferences not implemented in hybrid mode")
	} else if hasReadOwn {
		student, studentErr := s.getStudentByUserID(c)
		if studentErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to retrieve student profile: " + studentErr.Error()})
			return
		}

		// Implementasi s.achievementRepo.FindReferencesByStudentID(student.ID) diperlukan
		if student.ID == uuid.Nil {
			err = errors.New("student ID is nil, cannot fetch own achievements")
		}
		err = errors.New("FindReferencesByStudentID not implemented in hybrid mode")
	} else {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Access denied: insufficient permissions (need read_list or read_own)"})
		return
	}

	if err != nil && err.Error() != "FindReferencesByStudentID not implemented in hybrid mode" && err.Error() != "FindAllReferences not implemented in hybrid mode" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievements: " + err.Error()})
		return
	}

	// Mengembalikan 501 karena implementasi list hibrid kompleks dan belum selesai.
	c.JSON(http.StatusNotImplemented, gin.H{"status": "info", "message": "Listing achievements not fully implemented yet in hybrid mode. Repository functions need implementation."})
}

// @Summary Get Achievement Detail by ID
// @Description Mengambil detail prestasi berdasarkan ID.
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,data=model.AchievementDetailResponse}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /achievements/{id} [get]
func (s *achievementService) GetAchievementByID(c *gin.Context) {
	id := c.Param("id")
	achID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	// 1. Ambil Referensi PGSQL (Workflow)
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	// 2. Gabungkan data
	mergedAch, err := s.mergeAchievement(c.Request.Context(), pgRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": mergedAch})
}

// @Summary Update Achievement Data
// @Description Mahasiswa mengupdate data prestasi (hanya status DRAFT/REJECTED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Param achievement body model.AchievementUpdateRequest true "Data Prestasi yang Diperbarui"
// @Success 200 {object} object{status=string,data=model.AchievementDetailResponse}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /achievements/{id} [put]
func (s *achievementService) UpdateAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can update their achievements: " + err.Error()})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if pgRef.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner of this achievement"})
		return
	}

	// STATUS CHECK
	if pgRef.Status != model.StatusDraft && pgRef.Status != model.StatusRejected {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Cannot update achievement in status SUBMITTED or VERIFIED"})
		return
	}

	var req model.AchievementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body: " + err.Error()})
		return
	}

	// 2. SIAPKAN UPDATE UNTUK MONGODB (Data Konten)
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	// KONVERSI UPDATE DETAILS (STRUCT POINTER) ke map[string]interface{}
	if req.Details != nil {
		var detailsMap map[string]interface{}
		detailsBytes, err := json.Marshal(*req.Details) // Dereference pointer
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to process updated achievement details: " + err.Error()})
			return
		}

		if err := json.Unmarshal(detailsBytes, &detailsMap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Failed to parse updated achievement details: " + err.Error()})
			return
		}
		updates["details"] = detailsMap
	}

	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Points != nil {
		updates["points"] = *req.Points
	}
	updates["updatedAt"] = time.Now()

	// 3. UPDATE MONGODB
	if len(updates) > 1 {
		err = s.achievementRepo.UpdateMongoAchievement(c.Request.Context(), pgRef.MongoAchievementID, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update achievement content (Mongo): " + err.Error()})
			return
		}
	}

	// 4. UPDATE PGSQL (Workflow): Jika status Rejected, kembalikan ke Draft
	if pgRef.Status == model.StatusRejected {
		pgRef.Status = model.StatusDraft
		pgRef.RejectionNote = nil
	}
	pgRef.UpdatedAt = time.Now()

	if err := s.achievementRepo.UpdateReference(pgRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to update achievement reference (PGSQL): " + err.Error()})
		return
	}

	// Response gabungan
	mergedAch, _ := s.mergeAchievement(c.Request.Context(), pgRef)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement updated successfully", "data": mergedAch})
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
// @Router /achievements/{id} [delete]
func (s *achievementService) DeleteAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only students can delete their achievements: " + err.Error()})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if pgRef.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner of this achievement"})
		return
	}

	// STATUS CHECK
	if pgRef.Status != model.StatusDraft && pgRef.Status != model.StatusRejected {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Cannot delete achievement in status SUBMITTED or VERIFIED"})
		return
	}

	// 2. Hapus dari MongoDB (Hapus permanen content)
	if err := s.achievementRepo.DeleteMongoAchievement(c.Request.Context(), pgRef.MongoAchievementID); err != nil {
		// Log warning, tapi kita lanjutkan soft-delete referensi (PGSQL).
		fmt.Printf("Warning: Failed to delete achievement content (Mongo ID: %s): %v. Continuing with PGSQL delete.\n", pgRef.MongoAchievementID, err)
	}

	// 3. Hapus dari PostgreSQL (Soft Delete reference)
	if err := s.achievementRepo.DeleteReference(achID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to soft-delete achievement reference (PGSQL): " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement and associated content deleted successfully"})
}

// @Summary Submit Achievement for Verification
// @Description Mahasiswa mengubah status prestasi dari DRAFT/REJECTED menjadi SUBMITTED. File bukti wajib ada.
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,message=string}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /achievements/{id}/submit [post]
func (s *achievementService) SubmitAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Only students can submit: " + err.Error()})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	if pgRef.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner"})
		return
	}

	// STATUS CHECK
	if pgRef.Status != model.StatusDraft && pgRef.Status != model.StatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is already SUBMITTED or VERIFIED"})
		return
	}

	// PROOF CHECK (Cek Attachments di MongoDB)
	mongoAch, err := s.achievementRepo.FindMongoByID(c.Request.Context(), pgRef.MongoAchievementID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to retrieve achievement content for attachment check: " + err.Error()})
		return
	}
	if mongoAch == nil || len(mongoAch.Attachments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "At least one attachment must be uploaded before submitting"})
		return
	}

	// 2. Update Status dan Waktu di PGSQL (Pointer Manual)
	submitTime := time.Now()
	pgRef.Status = model.StatusSubmitted
	pgRef.SubmittedAt = &submitTime
	pgRef.UpdatedAt = time.Now()

	if err := s.achievementRepo.UpdateReference(pgRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to submit achievement (PGSQL): " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement submitted for verification"})
}

// @Summary Verify Achievement (Approve)
// @Description Dosen Wali/Admin menyetujui prestasi (Status SUBMITTED -> VERIFIED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} object{status=string,data=model.AchievementDetailResponse}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /achievements/{id}/verify [post]
func (s *achievementService) VerifyAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil || lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only authorized validators (Lecturers/Admins) can verify"})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	// STATUS CHECK
	if pgRef.Status != model.StatusSubmitted {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is not in SUBMITTED status"})
		return
	}

	// 2. Update Status, Validator ID, dan Waktu di PGSQL (Pointer Manual)
	verifyTime := time.Now()
	pgRef.Status = model.StatusVerified
	pgRef.VerifiedBy = &lecturer.ID
	pgRef.VerifiedAt = &verifyTime
	pgRef.UpdatedAt = time.Now()

	if err := s.achievementRepo.UpdateReference(pgRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to verify achievement (PGSQL): " + err.Error()})
		return
	}

	// Response gabungan
	mergedAch, _ := s.mergeAchievement(c.Request.Context(), pgRef)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement approved successfully", "data": mergedAch})
}

// @Summary Reject Achievement
// @Description Dosen Wali/Admin menolak prestasi (Status SUBMITTED -> REJECTED).
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Param rejection body model.AchievementRejectRequest true "Alasan Penolakan"
// @Success 200 {object} object{status=string,data=model.AchievementDetailResponse}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Router /achievements/{id}/reject [post]
func (s *achievementService) RejectAchievement(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	lecturer, err := s.lecturerRepo.FindByUserID(userUUID)
	if err != nil || lecturer == nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Only authorized validators (Lecturers/Admins) can reject"})
		return
	}

	var req model.AchievementRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body: " + err.Error()})
		return
	}

	if req.RejectionNote == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Rejection note is required"})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	// STATUS CHECK
	if pgRef.Status != model.StatusSubmitted {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Achievement is not in SUBMITTED status"})
		return
	}

	// 2. Update Status, Validator ID, dan Alasan Penolakan di PGSQL (Pointer Manual)
	rejectReason := req.RejectionNote
	pgRef.Status = model.StatusRejected
	pgRef.VerifiedBy = &lecturer.ID
	pgRef.RejectionNote = &rejectReason
	pgRef.UpdatedAt = time.Now()

	if err := s.achievementRepo.UpdateReference(pgRef); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to reject achievement (PGSQL): " + err.Error()})
		return
	}

	// Response gabungan
	mergedAch, _ := s.mergeAchievement(c.Request.Context(), pgRef)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Achievement rejected successfully", "data": mergedAch})
}

// @Summary Get Achievement History
// @Description Endpoint placeholder untuk melihat riwayat perubahan status prestasi.
// @Tags Achievements
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Achievement ID (UUID)"
// @Success 501 {object} object{status=string,message=string} "Not Implemented"
// @Router /achievements/{id}/history [get]
func (s *achievementService) GetAchievementHistory(c *gin.Context) {
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
// @Success 200 {object} object{status=string,message=string,data=model.AttachmentMetadata}
// @Failure 400 {object} object{status=string,message=string}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /achievements/{id}/attachments [post]
func (s *achievementService) UploadAttachment(c *gin.Context) {
	achID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid achievement ID format"})
		return
	}

	// 1. Ambil Referensi PGSQL
	pgRef, err := s.achievementRepo.FindReferenceByID(achID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch achievement reference: " + err.Error()})
		return
	}
	if pgRef == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Achievement not found"})
		return
	}

	// 2. Lakukan Ownership dan Status Check
	student, err := s.getStudentByUserID(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Only students can upload: " + err.Error()})
		return
	}

	if pgRef.StudentID != student.ID {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Not the owner"})
		return
	}

	if pgRef.Status != model.StatusDraft && pgRef.Status != model.StatusRejected {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Cannot upload attachment for achievement in status SUBMITTED or VERIFIED"})
		return
	}

	// 3. Ambil file dari form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File is required in 'file' form field"})
		return
	}

	// 4. Simpan file
	fileUUID := uuid.New().String()
	fileExtension := filepath.Ext(file.Filename)
	safeFilename := fmt.Sprintf("%s-%s%s",
		strings.ReplaceAll(achID.String(), "-", "")[:8],
		fileUUID,
		fileExtension)

	const uploadDir = "uploads/achievements"
	dst := filepath.Join(uploadDir, safeFilename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to create upload directory: " + err.Error()})
		return
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to save file: " + err.Error()})
		return
	}

	// 5. Update achievement.Attachments di MongoDB
	newAttachment := model.AttachmentMetadata{
		FileName:   file.Filename,
		FileUrl:    "/" + uploadDir + "/" + safeFilename,
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	mongoAch, err := s.achievementRepo.FindMongoByID(c.Request.Context(), pgRef.MongoAchievementID)
	if err != nil || mongoAch == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Data integrity error: Failed to retrieve Mongo content"})
		return
	}

	mongoAch.Attachments = append(mongoAch.Attachments, newAttachment)

	// UPDATE MONGODB
	updates := map[string]interface{}{
		"attachments": mongoAch.Attachments,
		"updatedAt":   time.Now(),
	}

	if err := s.achievementRepo.UpdateMongoAchievement(c.Request.Context(), pgRef.MongoAchievementID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to record file path in database (Mongo update failed): " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "File uploaded successfully", "data": newAttachment})
}
