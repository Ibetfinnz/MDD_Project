package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"auth"
)

// JWTMiddleware ใช้ตรวจสอบ JWT จาก header Authorization แล้วใส่ username, role ลงใน context
func JWTMiddleware(cfg *auth.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		// รองรับรูปแบบ "Bearer <token>"
		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header format"})
			return
		}

		tokenString := parts[1]

		claims, err := cfg.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// ใส่ข้อมูล user ลงใน Gin context เพื่อ handler ถัดไปนำไปใช้ได้
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
