package service

import (
	"net/http"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/fitrinovs/achievement_system/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(c *gin.Context)
	GetUserByID(c *gin.Context)
	GetUserByUsername(c *gin.Context) // Opsional, biasanya admin cari by ID
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser godoc
// @Summary      Create User (Admin)
// @Description  Membuat user baru (Admin/Dosen/Mahasiswa)
// @Tags         Users
// @Security     BearerAuth
// @Param        request body model.UserCreateRequest true "User Data"
// @Router       /api/v1/users [post]
func (s *userService) CreateUser(c *gin.Context) {
	var req model.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// 1. Cek Username Duplikat
	existingUser, _ := s.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "username already taken"})
		return
	}

	// 2. Cek Email Duplikat
	existingEmail, _ := s.userRepo.FindByEmail(req.Email)
	if existingEmail != nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "email already registered"})
		return
	}

	// 3. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to encrypt password"})
		return
	}

	// 4. Parse Role ID
	roleUUID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid role id format"})
		return
	}

	// 5. Setup Model User
	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: string(hashedPassword),
		RoleID:       roleUUID,
		IsActive:     true,
	}

	// 6. Simpan ke Database
	if err := s.userRepo.Create(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "data": newUser})
}

// GetUserByID godoc
// @Summary      Get User by ID
// @Tags         Users
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Router       /api/v1/users/{id} [get]
func (s *userService) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": user})
}

// GetUserByUsername (Opsional untuk internal/admin)
func (s *userService) GetUserByUsername(c *gin.Context) {
	username := c.Param("username") // Asumsi route /users/username/:username

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": user})
}

// UpdateUser godoc
// @Summary      Update User
// @Tags         Users
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Param        request body model.UserUpdateRequest true "Update Data"
// @Router       /api/v1/users/{id} [put]
func (s *userService) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	var req model.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Ambil data user lama
	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}

	// Update partial
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User updated successfully"})
}

// DeleteUser godoc
// @Summary      Delete User
// @Tags         Users
// @Security     BearerAuth
// @Param        id path string true "User UUID"
// @Router       /api/v1/users/{id} [delete]
func (s *userService) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid user id format"})
		return
	}

	if err := s.userRepo.Delete(userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted successfully"})
}
