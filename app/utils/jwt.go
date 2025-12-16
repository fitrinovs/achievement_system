package utils

import (
	"errors"
	"os"
	"time"

	"github.com/fitrinovs/achievement_system/app/model"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken membuat Access Token (short-lived: 1 jam)
func GenerateToken(user model.User, role string, permissions []string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	claims := model.JwtCustomClaims{
		UserID:      user.ID,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // 1 jam
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// GenerateRefreshToken membuat Refresh Token (long-lived: 7 hari)
func GenerateRefreshToken(user model.User, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	claims := model.JwtCustomClaims{
		UserID: user.ID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 hari
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateToken memvalidasi Access Token
func ValidateToken(tokenString string) (*model.JwtCustomClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &model.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.JwtCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken memvalidasi Refresh Token
func ValidateRefreshToken(tokenString string) (*model.JwtCustomClaims, error) {
	// Sama seperti ValidateToken, tapi khusus untuk refresh token
	return ValidateToken(tokenString)
}
