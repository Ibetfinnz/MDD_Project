package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Ibetfinnz/MDD_Project/auth"
	"github.com/gin-gonic/gin"
)

type CurrentUser struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AttachUserHeaders adds current user info into outgoing request headers
func AttachUserHeaders(c *gin.Context, req *http.Request) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return
	}

	req.Header.Set("X-User-Name", user.Username)
	req.Header.Set("X-User-Role", user.Role)
}

// GetCurrentUser reads the current user from gateway headers
func GetCurrentUser(c *gin.Context) (*CurrentUser, error) {
	username := c.GetHeader("X-User-Name")
	role := c.GetHeader("X-User-Role")
	if username == "" || role == "" {
		return nil, fmt.Errorf("not authorized")
	}

	return &CurrentUser{
		Username: username,
		Role:     role,
	}, nil
}

// RequireUser ensures the request has a logged-in user
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetCurrentUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "กรุณา login ก่อน",
			})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

// RequireAdmin allows only admin users
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := GetCurrentUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "กรุณา login ก่อน",
			})
			return
		}

		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "เฉพาะ admin เท่านั้น",
			})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

// JWTMiddleware validates JWT from Authorization header and sets user in context
func JWTMiddleware(cfg *auth.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization header format",
			})
			return
		}

		tokenString := parts[1]

		claims, err := cfg.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// 🔥 ใส่ลง context
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
