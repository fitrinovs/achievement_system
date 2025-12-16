package service

import (
	"context"
	"errors" 
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson" 
)

// =================================================================
// REPORT SERVICE INTERFACE (KONTRAK)
// =================================================================

type ReportService interface {
	GetAchievementStatistics(c *gin.Context)
	GetStudentAchievementReport(c *gin.Context)
}

// =================================================================
// REPORT SERVICE IMPLEMENTATION STRUCT
// =================================================================

type reportService struct {
	reportRepo      repository.ReportRepository
	studentRepo     repository.StudentRepository
	lecturerRepo    repository.LecturerRepository 
	achievementRepo repository.AchievementRepository 
}

// Constructor dengan semua dependency
func NewReportService(
	reportRepo repository.ReportRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	achievementRepo repository.AchievementRepository,
) ReportService {
	return &reportService{
		reportRepo: reportRepo,
		studentRepo: studentRepo,
		lecturerRepo: lecturerRepo,
		achievementRepo: achievementRepo,
	}
}

// =================================================================
// PRIVATE UTILITY FUNCTIONS
// =================================================================

// getStudentByUserID: Helper untuk mendapatkan Student UUID dari User UUID yang sedang login
func (s *reportService) getStudentByUserID(c *gin.Context) (*model.Student, error) {
	userIDStr := c.GetString("userID") 
	if userIDStr == "" {
		return nil, errors.New("user ID not found in context")
	}
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	
	// studentRepo.FindByUserID harus melakukan Preload("User")
	student, err := s.studentRepo.FindByUserID(userUUID) 
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}
	return student, nil
}

// processAggregationData melakukan perhitungan statistik dari data mentah MongoDB
func (s *reportService) processAggregationData(data []model.MongoAchievementMinimal) model.AchievementStatisticsResponse {
	response := model.AchievementStatisticsResponse{
		TotalAchievementCount: len(data),
		ByType:                []model.StatisticsByGroup{},
		ByLevel:               []model.StatisticsByGroup{},
		ByPeriod:              []model.StatisticsByGroup{},
	}

	typeCounts := make(map[string]int)
	levelCounts := make(map[string]int)
	periodCounts := make(map[string]int)

	for _, item := range data {
		// 1. By Type
		typeCounts[item.AchievementType]++

		// 2. By Level (menggunakan competitionLevel sebagai contoh level)
		if item.Level != "" {
			levelCounts[item.Level]++
		}

		// 3. By Period (Tahun Event)
		period := strconv.Itoa(item.EventDate.Year())
		if period != "" && period != "-1" { 
			periodCounts[period]++
		}
	}

	// Konversi map ke slice DTO
	for k, v := range typeCounts {
		response.ByType = append(response.ByType, model.StatisticsByGroup{Group: k, Count: v})
	}
	for k, v := range levelCounts {
		response.ByLevel = append(response.ByLevel, model.StatisticsByGroup{Group: k, Count: v})
	}
	for k, v := range periodCounts {
		response.ByPeriod = append(response.ByPeriod, model.StatisticsByGroup{Group: k, Count: v})
	}
	
	return response
}

// getFilterBasedOnRole menentukan filter agregasi MongoDB berdasarkan peran pengguna
func (s *reportService) getFilterBasedOnRole(c *gin.Context) (interface{}, error) {
	permissionsAny, _ := c.Get("permissions")
	permissions := permissionsAny.([]string)
	
	// Admin: Tidak ada filter (semua data), jika memiliki achievement:read_list
	for _, p := range permissions {
		if p == "achievement:read_list" { 
			return nil, nil 
		}
	}

	// Mahasiswa: Filter berdasarkan ID Mahasiswa sendiri (own)
	student, studentErr := s.getStudentByUserID(c) 
	if studentErr == nil {
		return bson.M{"studentId": student.ID}, nil
	}
	
	// Dosen Wali: Filter berdasarkan ID Mahasiswa bimbingan (advisee)
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)
	
	lecturer, lecturerErr := s.lecturerRepo.FindByUserID(userUUID) 
	if lecturerErr == nil {
		studentIDs, err := s.studentRepo.FindAdviseeIDsByAdvisorID(lecturer.ID) 
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve advisee IDs: %w", err)
		}
		if len(studentIDs) == 0 {
			return bson.M{"studentId": uuid.Nil}, nil 
		}
		return bson.M{"studentId": bson.M{"$in": studentIDs}}, nil
	}

	return nil, errors.New("access denied: user profile not recognized as student, admin, or lecturer")
}

// =================================================================
// PUBLIC HANDLER IMPLEMENTATION (DENGAN SWAGGER)
// =================================================================

// @Summary Generate Achievement Statistics
// @Description Menghasilkan statistik prestasi berdasarkan role pengguna (Admin: All, Dosen Wali: Advisee, Mahasiswa: Own). Sesuai FR-011.
// @Tags Reports
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} object{status=string,data=model.AchievementStatisticsResponse}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/reports/statistics [get]
func (s *reportService) GetAchievementStatistics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 1. Tentukan filter berdasarkan Role pengguna
	mongoFilter, err := s.getFilterBasedOnRole(c)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 2. Ambil data mentah dari MongoDB
	mongoData, err := s.reportRepo.AggregateAchievementData(ctx, mongoFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to aggregate achievement data"})
		return
	}

	// 3. Proses statistik (ByType, ByLevel, ByPeriod) dari data MongoDB
	statsResponse := s.processAggregationData(mongoData)
	
	// 4. Ambil Top Students (Hanya jika scope-nya Admin/Global)
	if mongoFilter == nil { 
		topStudents, err := s.reportRepo.GetTopStudentsByPoints(ctx, 10) 
		if err != nil {
			fmt.Printf("Warning: Failed to retrieve top students: %v\n", err)
		}
		statsResponse.TopStudents = topStudents
	} 

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": statsResponse})
}

// @Summary Get Student Achievement Report by ID
// @Description Mengambil detail semua prestasi (yang diverifikasi) dan total poin untuk mahasiswa tertentu (NIM atau UUID).
// @Tags Reports
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Student ID (UUID) atau Student ID (NIM)"
// @Success 200 {object} object{status=string,data=model.StudentAchievementReportResponse}
// @Failure 403 {object} object{status=string,message=string}
// @Failure 404 {object} object{status=string,message=string}
// @Failure 500 {object} object{status=string,message=string}
// @Router /api/v1/reports/student/{id} [get]
func (s *reportService) GetStudentAchievementReport(c *gin.Context) {
	studentIDParam := c.Param("id") // Ini bisa NIM atau UUID

	// 1. Cari profil mahasiswa berdasarkan UUID atau NIM
	var student *model.Student
	var err error
	
	studentUUID, uuidErr := uuid.Parse(studentIDParam)
	if uuidErr == nil {
		student, err = s.studentRepo.FindByID(studentUUID) // Cari berdasarkan UUID
	} else {
		// Jika bukan UUID, cari berdasarkan NIM. 
		// Method ini (FindByStudentID) di repo harus mencari berdasarkan kolom NIM.
		student, err = s.studentRepo.FindByStudentID(studentIDParam) 
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Student not found"})
		return
	}
	
	// 2. Otorisasi (Cek jika Mahasiswa, harus pemilik sendiri)
	userRole := c.GetString("role") 
	if userRole == "Mahasiswa" {
		loggedInStudent, _ := s.getStudentByUserID(c)
		if loggedInStudent == nil || loggedInStudent.ID != student.ID {
			c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "Forbidden: Cannot view other student's report"})
			return
		}
	}

	// 3. Ambil semua achievement mahasiswa tersebut
	achievements, err := s.achievementRepo.FindByStudentID(student.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch student achievements"})
		return
	}
	
	// 4. Hitung Total Poin (Hanya dari yang APPROVED)
	totalPoints := 0
	filteredAchievements := []model.Achievement{}
	for _, ach := range achievements {
		if ach.Status == model.StatusApproved { 
			totalPoints += ach.Points
			filteredAchievements = append(filteredAchievements, ach)
		}
	}
	
	// 5. Buat response
	report := model.StudentAchievementReportResponse{
		StudentProfile: &model.StudentProfileResponse{
			ID: student.ID,
			// FIX: Mengakses student.NIM
			NIM: student.NIM, 
			FullName: student.User.FullName, 
			ProgramStudy: student.ProgramStudy, // FIX: Menambahkan field baru
			AcademicYear: student.AcademicYear, // FIX: Menambahkan field baru
		},
		TotalPoints: totalPoints,
		Achievements: filteredAchievements,
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}