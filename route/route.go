package route

import (
	"github.com/fitrinovs/achievement_system/app/service"
	"github.com/fitrinovs/achievement_system/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// FIX: Menambahkan reportService sebagai parameter
func SetupRoutes(
	r *gin.Engine,
	authService service.AuthService,
	studentService service.StudentService,
	lecturerService service.LecturerService,
	userService service.UserService,
	achievementService service.AchievementService,
	reportService service.ReportService, // PARAMETER BARU: ReportService
) {

	// =========================
	// SWAGGER
	// =========================
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// =========================
	// AUTH (PUBLIC)
	// =========================
	authPublic := api.Group("/auth")
	{
		authPublic.POST("/login", authService.Login)
		authPublic.POST("/refresh", authService.RefreshToken)
	}

	// =========================
	// AUTHENTICATED ROUTES
	// =========================
	protected := api.Group("")
	// Middleware autentikasi
	protected.Use(middleware.AuthMiddleware()) 

	// =========================
	// AUTH (PROTECTED)
	// =========================
	authProtected := protected.Group("/auth")
	{
		authProtected.GET("/profile", authService.GetProfile)
		authProtected.POST("/logout", authService.Logout)
	}

	// =========================
	// USERS (ADMIN)
	// =========================
	userGroup := protected.Group("/users")
	{
		userGroup.GET("", userService.GetAllUsers)
		userGroup.POST("", userService.CreateUser)
		userGroup.GET("/:id", userService.GetUserByID)
		userGroup.PUT("/:id", userService.UpdateUser)
		userGroup.PUT("/:id/role", userService.UpdateUserRole)
		userGroup.DELETE("/:id", userService.DeleteUser)
	}

	// =================================================
	// STUDENTS
	// =================================================
	studentGroup := protected.Group("/students")
	{
		// CRUD
		studentGroup.POST("", studentService.CreateStudent)
		studentGroup.GET("", studentService.GetAllStudents)
		studentGroup.GET("/:id", studentService.GetStudentByID)
		studentGroup.PUT("/:id", studentService.UpdateStudent)
		studentGroup.DELETE("/:id", studentService.DeleteStudent)

		// SRS
		studentGroup.PUT("/:id/advisor", studentService.AssignAdvisor)
		studentGroup.GET("/:id/achievements", studentService.GetAchievementsByStudentID)
	}

	// =================================================
	// LECTURERS
	// =================================================
	lecturerGroup := protected.Group("/lecturers")
	{
		// CRUD
		lecturerGroup.POST("", lecturerService.CreateLecturer)
		lecturerGroup.GET("", lecturerService.GetAllLecturers)
		lecturerGroup.GET("/:id", lecturerService.GetLecturerByID)
		lecturerGroup.PUT("/:id", lecturerService.UpdateLecturer)
		lecturerGroup.DELETE("/:id", lecturerService.DeleteLecturer)

		// SRS
		lecturerGroup.GET("/:id/advisees", lecturerService.GetAdviseesByLecturerID)
	}

	// =================================================
	// ACHIEVEMENTS
	// =================================================
	achievementGroup := protected.Group("/achievements")
	{
		// 1. GET /api/v1/achievements (List)
		achievementGroup.GET("/", middleware.CheckPermission("achievement:read_own", "achievement:read_list"), achievementService.GetAchievementsList)

		// 2. GET /api/v1/achievements/:id (Detail)
		achievementGroup.GET("/:id", middleware.CheckPermission("achievement:read_own", "achievement:read_list"), achievementService.GetAchievementByID)

		// 3. POST /api/v1/achievements (Create) - Mahasiswa
		achievementGroup.POST("/", middleware.CheckPermission("achievement:create"), achievementService.CreateAchievement)

		// 4. PUT /api/v1/achievements/:id (Update) - Mahasiswa
		achievementGroup.PUT("/:id", middleware.CheckPermission("achievement:update"), achievementService.UpdateAchievement)

		// 5. DELETE /api/v1/achievements/:id (Delete) - Mahasiswa
		achievementGroup.DELETE("/:id", middleware.CheckPermission("achievement:delete"), achievementService.DeleteAchievement)

		// 6. POST /api/v1/achievements/:id/submit - Mahasiswa
		achievementGroup.POST("/:id/submit", middleware.CheckPermission("achievement:submit"), achievementService.SubmitAchievement)

		// 7. POST /api/v1/achievements/:id/verify - Dosen Wali/Admin
		achievementGroup.POST("/:id/verify", middleware.CheckPermission("achievement:verify"), achievementService.VerifyAchievement)

		// 8. POST /api/v1/achievements/:id/reject - Dosen Wali/Admin
		achievementGroup.POST("/:id/reject", middleware.CheckPermission("achievement:reject"), achievementService.RejectAchievement)

		// 9. GET /api/v1/achievements/:id/history
		achievementGroup.GET("/:id/history", middleware.CheckPermission("achievement:read_history"), achievementService.GetAchievementHistory)
		
		// 10. POST /api/v1/achievements/:id/attachments - Mahasiswa
		achievementGroup.POST("/:id/attachments", middleware.CheckPermission("achievement:upload_attachment"), achievementService.UploadAttachment)
	}

	// =================================================
	// 8. REPORTS & ANALYTICS
	// =================================================
	SetupReportRoutes(protected, reportService)
}

// Helper function untuk mendaftarkan route report di grup protected
func SetupReportRoutes(router *gin.RouterGroup, reportService service.ReportService) {
	reports := router.Group("/reports")
	{
		// GET /api/v1/reports/statistics
		reports.GET("/statistics", reportService.GetAchievementStatistics)
		
		// GET /api/v1/reports/student/:id
		reports.GET("/student/:id", reportService.GetStudentAchievementReport)
	}
}