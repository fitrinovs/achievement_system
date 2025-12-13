package middleware

import (
	"net/http"
	"strings"

	"github.com/fitrinovs/achievement_system/app/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware memeriksa validitas JWT token di header Authorization
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Validasi token menggunakan utils yang sudah kita buat
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Simpan data user (claims) ke context agar bisa dipakai di route selanjutnya
		// Ini berguna untuk mengecek permission nanti
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// PermissionMiddleware mengecek apakah user memiliki permission tertentu
// Referensi SRS: FR-002 Flow 4 "Check apakah user memiliki permission"
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil permissions dari context (yang diset oleh AuthMiddleware)
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permissions not found"})
			c.Abort()
			return
		}

		userPermissions, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permission format"})
			c.Abort()
			return
		}

		// Cek apakah requiredPermission ada di dalam list userPermissions
		hasPermission := false
		for _, p := range userPermissions {
			if p == requiredPermission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}
