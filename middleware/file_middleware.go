package middleware

import (
	"github.com/fitrinovs/achievement_system/app/utils"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			utils.ErrorResponse(c, 403, "No permissions found", nil)
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			utils.ErrorResponse(c, 403, "Invalid permissions format", nil)
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission := false
		for _, p := range permissionList {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.ErrorResponse(c, 403, "Insufficient permissions", map[string]string{
				"required": permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, 403, "No role found", nil)
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			utils.ErrorResponse(c, 403, "Invalid role format", nil)
			c.Abort()
			return
		}

		// Check if user has one of required roles
		hasRole := false
		for _, r := range roles {
			if r == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.ErrorResponse(c, 403, "Insufficient role privileges", map[string]interface{}{
				"required": roles,
				"current":  role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
