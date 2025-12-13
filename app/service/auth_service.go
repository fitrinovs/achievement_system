package service

import (
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/fitrinovs/achievement_system/app/utils" // Pastikan import ini mengarah ke utils

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(c *gin.Context)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
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
// @Router       /api/v1/auth/login [post]
func (s *authService) Login(c *gin.Context) {
	// 1. TERIMA REQUEST (HTTP Parsing)
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 2. LOGIKA BISNIS
	// Cari user berdasarkan username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid username or password"})
		return
	}

	// Cek Password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid username or password"})
		return
	}

	// Ambil Permission user
	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to fetch user permissions"})
		return
	}

	// Cek Role Name
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	// 3. GENERATE TOKEN (Panggil UTILS)
	// Perbaikan: Menggunakan 'utils.GenerateToken', bukan 'utils.GenerateToken'
	tokenString, err := utils.GenerateToken(*user, roleName, permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate token"})
		return
	}

	// Siapkan Data Response
	response := &model.LoginResponse{
		Token:        tokenString,
		RefreshToken: "", // Kosongkan jika belum implementasi refresh token
		User: model.UserProfileResponse{
			ID:          user.ID,
			Username:    user.Username,
			Fullname:    user.FullName,
			Email:       user.Email,
			Role:        roleName,
			Permissions: permissions,
		},
	}

	// 4. KIRIM RESPONSE JSON
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": response})
}
