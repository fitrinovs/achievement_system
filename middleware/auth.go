// File: middleware/auth.go

package middleware

import (
	"fmt" // DIBUTUHKAN untuk error message yang lebih baik
	"net/http"
	"strings"

	"github.com/fitrinovs/achievement_system/app/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware memeriksa validitas JWT token di header Authorization
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// PERBAIKAN: Gunakan "userID" agar konsisten dengan c.GetString("userID") di service layer
		c.Set("userID", claims.UserID) 
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// CheckPermission mengecek apakah user memiliki SALAH SATU (OR logic) dari permissions yang diberikan.
// Ini adalah function yang Anda panggil di route.go
func CheckPermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil permissions dari context (yang diset oleh AuthMiddleware)
		permissionsAny, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permissions not found in context (Run AuthMiddleware first)"})
			c.Abort()
			return
		}

		userPermissions, ok := permissionsAny.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permission format in context"})
			c.Abort()
			return
		}

		// Cek apakah user memiliki SALAH SATU dari permissions yang disyaratkan
		hasPermission := false
		for _, reqP := range requiredPermissions {
			for _, userP := range userPermissions {
				if userP == reqP {
					hasPermission = true
					break // Keluar dari inner loop (userP)
				}
			}
			if hasPermission {
				break // Keluar dari outer loop (reqP)
			}
		}

		if !hasPermission {
			// Pesan error yang lebih informatif
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "You don't have permission to access this resource",
				"details": fmt.Sprintf("Required: %s", strings.Join(requiredPermissions, " OR ")),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Catatan: PermissionMiddleware yang lama sudah tidak diperlukan 
// karena CheckPermission(requiredPermission string) sudah bisa menggantikannya.