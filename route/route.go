package route

import (
	"github.com/fitrinovs/achievement_system/app/service"
	"github.com/fitrinovs/achievement_system/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// PERBAIKAN: Fungsi ini sudah memiliki 6 parameter Service
func SetupRoutes(
	r *gin.Engine,
	authService service.AuthService,
	studentService service.StudentService,
	lecturerService service.LecturerService,
	userService service.UserService, // <-- UserService sudah ada
	achievementService service.AchievementService,
) {
	// Integrasi Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// ==========================================
	// 1. AUTHENTICATION (Public Routes)
	// ==========================================

	authPublicGroup := api.Group("/auth")
	{
		authPublicGroup.POST("/login", authService.Login)
		authPublicGroup.POST("/refresh", authService.RefreshToken)
	}

	// ==========================================
	// 2. PROTECTED ROUTES
	// ==========================================

	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// --- AUTH PROTECTED ROUTES ---
	authProtectedGroup := protected.Group("/auth")
	{
		authProtectedGroup.GET("/profile", authService.GetProfile)
		authProtectedGroup.POST("/logout", authService.Logout)
	}

	// --- USER ROUTES --- (PROTECTED)
	userGroup := protected.Group("/users")
	{
		userGroup.GET("", userService.GetAllUsers)      // TAMBAHAN: Get All Users
		userGroup.POST("", userService.CreateUser)
		userGroup.GET("/:id", userService.GetUserByID)
		userGroup.PUT("/:id", userService.UpdateUser)
		userGroup.PUT("/:id/role", userService.UpdateUserRole) // TAMBAHAN: Update User Role
		userGroup.DELETE("/:id", userService.DeleteUser)
	}

	// --- STUDENT ROUTES ---
	studentGroup := protected.Group("/students")
	{
		studentGroup.POST("", studentService.CreateStudent)
		studentGroup.GET("/:id", studentService.GetStudentByID)
		studentGroup.PUT("/:id", studentService.UpdateStudent)
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
		achievementGroup.POST("", achievementService.CreateAchievement)
		achievementGroup.GET("/my", achievementService.GetMyAchievements)
		achievementGroup.GET("/:id", achievementService.GetAchievementByID)

		achievementGroup.PUT("/:id/validate", achievementService.ValidateAchievement)
	}

}