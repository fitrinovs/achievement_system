package route

import (
	"github.com/fitrinovs/achievement_system/app/service"
	"github.com/fitrinovs/achievement_system/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(
	r *gin.Engine,
	authService service.AuthService,
	studentService service.StudentService,
	lecturerService service.LecturerService,
	achievementService service.AchievementService,
) {
	// Integrasi Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// ==========================================
	// 1. AUTHENTICATION (Public Routes)
	// ==========================================

	// Langsung panggil nama fungsi service-nya
	api.POST("/auth/login", authService.Login)

	// ==========================================
	// 2. PROTECTED ROUTES
	// ==========================================

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// --- STUDENT ROUTES ---
	// Perbaikan: Hapus wrapper func(), langsung panggil service
	studentGroup := protected.Group("/students")
	{
		studentGroup.POST("", studentService.CreateStudent)
		studentGroup.GET("/:id", studentService.GetStudentByID)
		studentGroup.PUT("/:id", studentService.UpdateStudent)
		// studentGroup.DELETE("/:id", studentService.DeleteStudent) // Jika ada
	}

	// --- LECTURER ROUTES ---
	lecturerGroup := protected.Group("/lecturers")
	{
		lecturerGroup.POST("", lecturerService.CreateLecturer)
		lecturerGroup.GET("", lecturerService.GetAllLecturers)
		lecturerGroup.GET("/:id", lecturerService.GetLecturerByID)
		lecturerGroup.GET("/user/:user_id", lecturerService.GetLecturerByUserID)
		lecturerGroup.PUT("/:id", lecturerService.UpdateLecturer)
		lecturerGroup.DELETE("/:id", lecturerService.DeleteLecturer)
	}

	achievementGroup := protected.Group("/achievements")
	{
		achievementGroup.POST("", achievementService.CreateAchievement)     // Upload
		achievementGroup.GET("/my", achievementService.GetMyAchievements)   // List Punya Sendiri
		achievementGroup.GET("/:id", achievementService.GetAchievementByID) // Detail

		// Validasi (Bisa tambahkan middleware cek role dosen jika mau lebih strict)
		achievementGroup.PUT("/:id/validate", achievementService.ValidateAchievement)
	}

}
