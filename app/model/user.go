package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"type:varchar(50);unique;not null"`
	Email        string    `json:"email" gorm:"type:varchar(100);unique;not null"`
	FullName     string    `json:"full_name" gorm:"type:varchar(255);not null"`
	PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`

	RoleID uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	Role   *Role     `json:"role,omitempty" gorm:"foreignKey:RoleID"`

	IsActive bool `json:"is_active" gorm:"default:true"`

	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// kredensial login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// respon login sukses
type LoginResponse struct {
	Token        string              `json:"token"`
	RefreshToken string              `json:"refresh_token"`
	User         UserProfileResponse `json:"user"`
}

type UserProfileResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Fullname    string    `json:"fullname"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
}

type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name"`
	// Ubah int menjadi string agar kompatibel dengan UUID
	RoleID string `json:"role_id" binding:"required"`
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	// Ubah int menjadi string
	RoleID string `json:"role_id"`
}
