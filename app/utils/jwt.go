package utils

import (
	"errors"
	"os"
	"time"

	"github.com/fitrinovs/achievement_system/app/model" // Pastikan ini sesuai go.mod Anda

	"github.com/golang-jwt/jwt/v5"
)

// Gunakan secret key dari environment variable
var jwtSecret = []byte(getEnv("JWT_SECRET", "RahasiaNegara123!"))

// GenerateToken membuat token JWT baru
// Input user menggunakan model.User yang ID-nya sudah uuid.UUID
func GenerateToken(user model.User, roleName string, permissions []string) (string, error) {
	// Set waktu kadaluarsa (24 jam)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Buat claims
	claims := &model.JwtCustomClaims{
		UserID:      user.ID, // Ini sekarang tipe uuid.UUID, sudah cocok
		Role:        roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sistem-prestasi-mahasiswa",
		},
	}

	// Create token dengan algoritma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token dengan secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken memvalidasi token string dan mengembalikan claims jika valid
func ValidateToken(tokenString string) (*model.JwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validasi algoritma signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*model.JwtCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
