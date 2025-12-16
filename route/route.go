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
	userService service.UserService,
	achievementService service.AchievementService,
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
	// STUDENTS (SESUAI SRS & STUDENT SERVICE)
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
	// LECTURERS (SESUAI SRS)
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
		achievementGroup.POST("", achievementService.CreateAchievement)
		achievementGroup.GET("/my", achievementService.GetMyAchievements)
		achievementGroup.GET("/:id", achievementService.GetAchievementByID)

		// Verifikasi (Dosen Wali)
		achievementGroup.PUT("/:id/validate", achievementService.ValidateAchievement)
	}
}
