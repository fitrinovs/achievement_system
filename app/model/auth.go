package model

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JwtCustomClaims adalah struktur data yang akan disimpan di dalam payload Token (JWT).
// Kita sesuaikan UserID menjadi uuid.UUID agar sama dengan model User.
type JwtCustomClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}
