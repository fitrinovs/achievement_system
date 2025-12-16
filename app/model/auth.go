package model

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtCustomClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	Token string `json:"token"` 
}

