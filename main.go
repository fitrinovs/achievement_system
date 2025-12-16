package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/fitrinovs/achievement_system/app/service"
	"github.com/fitrinovs/achievement_system/config"
	"github.com/fitrinovs/achievement_system/database"
	"github.com/fitrinovs/achievement_system/middleware"
	"github.com/fitrinovs/achievement_system/route"

	_ "github.com/fitrinovs/achievement_system/docs"

	"github.com/gin-gonic/gin"
)

// @title Student Achievement API
// @version 1.0
// @description API untuk Sistem Pelaporan Prestasi Mahasiswa
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@unair.ac.id

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Setup logger
	logger := config.SetupLogger()
	logger.Info("🚀 Starting Student Achievement System...")

	// 3. Connect to PostgreSQL
	logger.Info("Connecting to PostgreSQL...")
	database.ConnectPostgreSQL(cfg)
	logger.Info("✅ PostgreSQL connected successfully!")

	// 4. Connect to MongoDB
	logger.Info("Connecting to MongoDB...")
	database.ConnectMongoDB(cfg)
	defer database.DisconnectMongoDB()
	logger.Info("✅ MongoDB connected successfully!")

	// 5. Auto migrate database tables
	logger.Info("Running database migrations...")
	database.MigrateDatabase(
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.User{},
		&model.Lecturer{},
		&model.Student{},
		&model.Achievement{},
	)
	logger.Info("✅ Database migration completed!")

	// 6. Create uploads directory
	uploadPath := cfg.Upload.Path
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.Fatal("❌ Failed to create uploads directory:", err)
	}
	// Buat subfolder achievement
	if err := os.MkdirAll(uploadPath+"/achievements", 0755); err != nil {
		log.Fatal("❌ Failed to create uploads/achievements directory:", err)
	}

	// ========================================================
	// 7. INIT LAYERS (Dependency Injection)
	// ========================================================

	// A. Repositories
	userRepo := repository.NewUserRepository(database.DB)
	lecturerRepo := repository.NewLecturerRepository(database.DB)
	studentRepo := repository.NewStudentRepository(database.DB)
	achievementRepo := repository.NewAchievementRepository(database.DB)

	// B. Services
	// PERBAIKAN: NewAuthService membutuhkan userRepo, studentRepo, dan lecturerRepo
	authService := service.NewAuthService(userRepo, studentRepo, lecturerRepo)
	studentService := service.NewStudentService(studentRepo, userRepo, lecturerRepo)
	lecturerService := service.NewLecturerService(lecturerRepo, userRepo)
	// TAMBAHAN: Inisialisasi UserService
	userService := service.NewUserService(userRepo) 

	achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo)

	// ========================================================
	// 8. SETUP GIN SERVER
	// ========================================================
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Panic Recovery Middleware
	router.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Println("🔥🔥🔥 PANIC DETECTED 🔥🔥🔥")
				log.Println(r)
			}
		}()
		c.Next()
	})

	// Apply middlewares
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", healthCheckHandler)

	// Root Route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": 	     "Welcome to Student Achievement System API",
			"version": 	     "1.0",
			"documentation": "/swagger/index.html",
		})
	})

	// ========================================================
	// 9. SETUP ROUTES
	// ========================================================

	// PERBAIKAN: Tambahkan userService sebagai parameter
	route.SetupRoutes(router, authService, studentService, lecturerService, userService, achievementService)

	// Serve static files
	router.Static("/uploads", uploadPath)

	// 10. Start server
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info(fmt.Sprintf("🌐 Server is running on http://localhost%s", port))
	logger.Info(fmt.Sprintf("📄 Swagger Docs available at http://localhost%s/swagger/index.html", port))

	if err := router.Run(port); err != nil {
		logger.Fatal(fmt.Sprintf("❌ Failed to start server: %v", err))
	}
}

// ... (func corsMiddleware dan healthCheckHandler tetap sama) ...
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func healthCheckHandler(c *gin.Context) {
	sqlDB, err := database.DB.DB()
	postgresStatus := "connected"
	if err != nil || sqlDB.Ping() != nil {
		postgresStatus = "disconnected"
	}

	mongoStatus := "connected"
	if database.MongoDB == nil {
		mongoStatus = "disconnected"
	}

	c.JSON(200, gin.H{
		"status": 	"ok",
		"service": "Student Achievement System",
		"database": gin.H{
			"postgresql": postgresStatus,
			"mongodb": 	mongoStatus,
		},
	})
}