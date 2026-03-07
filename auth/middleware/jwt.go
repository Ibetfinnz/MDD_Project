package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/Ibetfinnz/MDD_Project/auth"
)

type CurrentUser struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// AttachUserHeaders แนบข้อมูล user ปัจจุบันลงใน request สำหรับเรียก service อื่น
// อ่านข้อมูลจาก GetCurrentUser เพียงจุดเดียว เพื่อไม่ต้องไป get header ซ้ำในทุก service
func AttachUserHeaders(c *gin.Context, req *http.Request) {
	user, err := GetCurrentUser(c)
	if err != nil {
		return
	}

	req.Header.Set("X-User-Name", user.Username)
	req.Header.Set("X-User-Role", user.Role)
}

// GetCurrentUser อ่าน user จาก gin context
func GetCurrentUser(c *gin.Context) (*CurrentUser, error) {
	// 1) ลองอ่านจาก Gin context ก่อน (กรณีผ่าน JWTMiddleware)
	if usernameVal, userExists := c.Get("username"); userExists {
		if roleVal, roleExists := c.Get("role"); roleExists {
			username, ok1 := usernameVal.(string)
			role, ok2 := roleVal.(string)
			if ok1 && ok2 {
				return &CurrentUser{Username: username, Role: role}, nil
			}
		}
	}

	// 2) fallback: อ่านจาก header ที่ Gateway ใส่ให้ (X-User-Name / X-User-Role)
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

// RequireUser เป็น Gin middleware ที่บังคับให้ต้องมี user (ต้อง login แล้ว)
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

// RequireAdmin เป็น Gin middleware ที่บังคับให้ต้องเป็น admin เท่านั้น
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

// JWTMiddleware ใช้ตรวจสอบ JWT จาก header Authorization แล้วใส่ username, role ลงใน context
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
