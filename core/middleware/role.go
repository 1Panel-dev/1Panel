package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/gin-gonic/gin"
)

var userService = service.NewIUserService()

// RoleAuth middleware checks if user has required role
func RoleAuth(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session user
		sessionData, err := global.SESSION.Get(c)
		if err != nil {
			helper.BadAuth(c, "ErrNotLogin", buserr.New("session not found"))
			return
		}

		sessionUser, ok := sessionData.(*psession.SessionUser)
		if !ok {
			helper.BadAuth(c, "ErrNotLogin", buserr.New("invalid session data"))
			return
		}

		// Get user from database
		user, err := userService.GetUserByUsername(sessionUser.Name)
		if err != nil {
			helper.BadAuth(c, "ErrUserNotFound", buserr.New("user not found"))
			return
		}

		// Check if user role is allowed
		allowed := false
		for _, role := range requiredRoles {
			if user.Role == role {
				allowed = true
				break
			}
		}

		if !allowed && user.Role != constant.RoleAdminMain {
			helper.BadAuth(c, "ErrUnauthorized", buserr.New("insufficient permissions"))
			return
		}

		// Set user info in context
		c.Set("UserID", user.ID)
		c.Set("UserRole", user.Role)
		c.Set("Username", user.Username)

		// Get and set permissions
		permissions, _ := userService.GetUserPermissions(user.ID)
		c.Set("Permissions", permissions)

		c.Next()
	}
}

// PermissionAuth middleware checks if user has required permission
func PermissionAuth(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session user
		sessionData, err := global.SESSION.Get(c)
		if err != nil {
			helper.BadAuth(c, "ErrNotLogin", buserr.New("session not found"))
			return
		}

		sessionUser, ok := sessionData.(*psession.SessionUser)
		if !ok {
			helper.BadAuth(c, "ErrNotLogin", buserr.New("invalid session data"))
			return
		}

		// Get user from database
		user, err := userService.GetUserByUsername(sessionUser.Name)
		if err != nil {
			helper.BadAuth(c, "ErrUserNotFound", buserr.New("user not found"))
			return
		}

		// Admin has all permissions
		if user.Role == constant.RoleAdminMain {
			c.Set("UserID", user.ID)
			c.Set("UserRole", user.Role)
			c.Set("Username", user.Username)
			c.Next()
			return
		}

		// Check if user has required permissions
		permissions, _ := userService.GetUserPermissions(user.ID)
		hasPermission := false

		for _, requiredPerm := range requiredPermissions {
			for _, userPerm := range permissions {
				if userPerm == requiredPerm || userPerm == constant.PermissionAdminAll {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("UserID", user.ID)
		c.Set("UserRole", user.Role)
		c.Set("Username", user.Username)
		c.Set("Permissions", permissions)

		c.Next()
	}
}

// UserAuthMiddleware adds user information to context
func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for auth endpoints
		if strings.Contains(c.Request.URL.Path, "/api/v2/core/auth") {
			c.Next()
			return
		}

		// Get session user
		sessionData, err := global.SESSION.Get(c)
		if err != nil {
			c.Next()
			return
		}

		sessionUser, ok := sessionData.(*psession.SessionUser)
		if !ok {
			c.Next()
			return
		}

		// Get user from database
		user, err := userService.GetUserByUsername(sessionUser.Name)
		if err != nil {
			c.Next()
			return
		}

		// Set user info in context
		c.Set("UserID", user.ID)
		c.Set("UserRole", user.Role)
		c.Set("Username", user.Username)

		// Get and set permissions
		permissions, _ := userService.GetUserPermissions(user.ID)
		c.Set("Permissions", permissions)

		c.Next()
	}
}
