package service

import (
	"net/http"
	"strings"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/fitrinovs/achievement_system/app/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(c *gin.Context)
	GetProfile(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
}

type authService struct {
	userRepo     repository.UserRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

// Login godoc
// @Summary      Login User
// @Description  Masuk sistem untuk mendapatkan Token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "Username & Password"
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/login [post]
func (s *authService) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid username or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid username or password"})
		return
	}

	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to fetch user permissions"})
		return
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	tokenString, err := utils.GenerateToken(*user, roleName, permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(*user, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate refresh token"})
		return
	}

	response := &model.LoginResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
		User: model.UserProfileResponse{
			ID:          user.ID,
			Username:    user.Username,
			Fullname:    user.FullName,
			Email:       user.Email,
			Role:        roleName,
			Permissions: permissions,
		},
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": response})
}

// GetProfile godoc
// @Summary      Get User Profile
// @Description  Mendapatkan profile user yang sedang login
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{status=string,data=model.UserProfileResponse}
// @Failure      401  {object}  object{status=string,message=string}
// @Failure      404  {object}  object{status=string,message=string}
// @Router       /auth/profile [get]
func (s *authService) GetProfile(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid user session"})
		return
	}

	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		permissions = []string{}
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	// Tambahan: Ambil data Student atau Lecturer jika ada
	var studentData *model.Student
	var lecturerData *model.Lecturer

	if roleName == "Student" || roleName == "Mahasiswa" {
		studentData, _ = s.studentRepo.FindByUserID(user.ID)
	} else if roleName == "Lecturer" || roleName == "Dosen" {
		lecturerData, _ = s.lecturerRepo.FindByUserID(user.ID)
	}

	profile := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"fullname":    user.FullName,
		"email":       user.Email,
		"role":        roleName,
		"permissions": permissions,
		"is_active":   user.IsActive,
		"created_at":  user.CreatedAt,
	}

	if studentData != nil {
		profile["student"] = studentData
	}

	if lecturerData != nil {
		profile["lecturer"] = lecturerData
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": profile})
}


// RefreshToken godoc
// @Summary      Refresh Access Token
// @Description  Mendapatkan access token baru menggunakan refresh token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body model.RefreshTokenRequest true "Refresh Token"
// @Success      200  {object}  object{status=string,data=model.RefreshTokenResponse}
// @Failure      400  {object}  object{status=string,message=string}
// @Failure      401  {object}  object{status=string,message=string}
// @Router       /auth/refresh [post]
func (s *authService) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid or expired refresh token"})
		return
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "user not found"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "user is inactive"})
		return
	}

	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		permissions = []string{}
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	newToken, err := utils.GenerateToken(*user, roleName, permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate new token"})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(*user, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate new refresh token"})
		return
	}

	response := model.RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": response})
}

// Logout godoc
// @Summary      Logout User
// @Description  Logout dari sistem (client-side token removal)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  object{status=string,message=string}
// @Router       /auth/logout [post]
func (s *authService) Logout(c *gin.Context) {
	// Ambil token dari header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "no token provided"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// OPTIONAL: Implementasi Token Blacklist
	// Untuk production, simpan token ke Redis/Database dengan expiry time
	// Contoh: utils.BlacklistToken(tokenString, expiryTime)

	// NOTE: Untuk sekarang, logout hanya mengandalkan client menghapus token
	// Server tidak perlu melakukan apa-apa karena JWT adalah stateless
	// Client bertanggung jawab untuk menghapus token dari localStorage/cookie

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "logged out successfully. please remove token from client",
		"token":   tokenString[:20] + "...", // Log sebagian token untuk tracking
	})
}
