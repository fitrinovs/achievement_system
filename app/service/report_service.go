// File: app/service/report_service.go (FINAL REVISI UNTUK MENGHILANGKAN ERROR ROUTER)

package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =================================================================
// REPORT SERVICE INTERFACE (DITAMBAHKAN METHOD YANG HILANG)
// =================================================================

type ReportService interface {
	GetReportByStudentID(c *gin.Context)
	
	// METHOD YANG DIBUTUHKAN OLEH ROUTER (ERROR 1)
	GetAchievementStatistics(c *gin.Context) 
	
	// METHOD YANG DIBUTUHKAN OLEH ROUTER (ERROR 2)
	GetStudentAchievementReport(c *gin.Context) 
}

// =================================================================
// REPORT SERVICE IMPLEMENTATION STRUCT & CONSTRUCTOR
// =================================================================

type reportService struct {
	achievementRepo repository.AchievementRepository
	studentRepo     repository.StudentRepository
}

func NewReportService(
	achievementRepo repository.AchievementRepository,
	studentRepo repository.StudentRepository,
) ReportService {
	return &reportService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
	}
}

// =================================================================
// PRIVATE UTILITY (Helper untuk menggabungkan data hibrid)
// =================================================================

// mergeAchievement menggabungkan data referensi PGSQL dengan konten MongoDB
func (s *reportService) mergeAchievement(ctx context.Context, pgRef *model.AchievementReference) (*model.AchievementDetailResponse, error) {
	mongoAch, err := s.achievementRepo.FindMongoByID(ctx, pgRef.MongoAchievementID)
	if err != nil {
		return nil, errors.New("failed to find achievement content in MongoDB")
	}
	if mongoAch == nil {
		return nil, errors.New("data integrity error: mongo content missing")
	}
	
	return &model.AchievementDetailResponse{
		ID:            pgRef.ID,
		StudentID:     pgRef.StudentID,
		Status:        pgRef.Status,
		SubmittedAt:   pgRef.SubmittedAt,
		VerifiedAt:    pgRef.VerifiedAt,
		VerifiedBy:    pgRef.VerifiedBy,
		RejectionNote: pgRef.RejectionNote,
		
		MongoAchievementID: pgRef.MongoAchievementID,
		AchievementType:    mongoAch.AchievementType,
		Title:              mongoAch.Title,
		Description:        mongoAch.Description,
		Details:            mongoAch.Details,
		Attachments:        mongoAch.Attachments,
		Tags:               mongoAch.Tags,
		Points:             mongoAch.Points,
		
		CreatedAt:          pgRef.CreatedAt, 
		UpdatedAt:          pgRef.UpdatedAt,
	}, nil
}


// =================================================================
// PUBLIC HANDLER IMPLEMENTATION
// =================================================================

// IMPLEMENTASI LAMA (SUDAH BENAR)
func (s *reportService) GetReportByStudentID(c *gin.Context) {
	studentIDStr := c.Param("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid student id format"})
		return
	}

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil || student == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "student not found"})
		return
	}

	achievementsRef, err := s.achievementRepo.FindReferencesByStudentID(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to retrieve achievements references"})
		return
	}

	totalPoints := 0
	var verifiedAchievements []model.AchievementDetailResponse

	for _, achRef := range achievementsRef {
		if achRef.Status == model.StatusVerified { 
			mergedAch, mergeErr := s.mergeAchievement(c.Request.Context(), &achRef)
			
			if mergeErr == nil {
				totalPoints += mergedAch.Points
				verifiedAchievements = append(verifiedAchievements, *mergedAch)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"student_id":            student.ID,
			"student_name":          student.NIM, 
			"total_verified_points": totalPoints,
			"verified_achievements": verifiedAchievements,
		},
	})
}

// =================================================================
// IMPLEMENTASI BARU UNTUK MENGHILANGKAN ERROR ROUTER
// =================================================================

// @Summary Get Global Achievement Statistics
// @Description Endpoint placeholder untuk menampilkan statistik prestasi global.
// @Tags Reports
// @Security BearerAuth
// @Success 501 {object} object{status=string,message=string}
// @Router /api/v1/reports/statistics [get]
func (s *reportService) GetAchievementStatistics(c *gin.Context) {
	// Pemasangan sementara (Stub) agar router tidak error.
	c.JSON(http.StatusNotImplemented, gin.H{"status": "info", "message": "Endpoint GetAchievementStatistics belum diimplementasikan."})
}

// @Summary Get Student Achievement Report (List)
// @Description Endpoint placeholder untuk mendapatkan daftar laporan prestasi mahasiswa (Mungkin sama dengan GetReportByStudentID jika ID dihilangkan).
// @Tags Reports
// @Security BearerAuth
// @Success 501 {object} object{status=string,message=string}
// @Router /api/v1/reports/student-report [get]
func (s *reportService) GetStudentAchievementReport(c *gin.Context) {
	// Pemasangan sementara (Stub) agar router tidak error.
	c.JSON(http.StatusNotImplemented, gin.H{"status": "info", "message": "Endpoint GetStudentAchievementReport belum diimplementasikan."})
}