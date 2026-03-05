package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Ibetfinnz/MDD_Project/auth"
	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *gorm.DB
var authConfig *auth.AuthConfig

// ===== Connect DB =====
func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("user.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connect error:", err)
	}

	db.AutoMigrate(&User{})

	// seed user
	db.FirstOrCreate(&User{}, User{
		Username: "admin",
		Password: "1234",
		Role:     "admin",
	})

	db.FirstOrCreate(&User{}, User{
		Username: "tenant1",
		Password: "1234",
		Role:     "tenant",
	})
}

// ===== Handlers =====

// Service check
func serviceCheck(c *gin.Context) {
	c.String(http.StatusOK, "User Service Running")
}

// Get all users
func getAllUsers(c *gin.Context) {
	var users []User
	if err := db.Select("username, role").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Login
func login(c *gin.Context) {
	var input LoginInput
	c.ShouldBindJSON(&input)

	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login"})
		return
	}

	if user.Password != input.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString(middleware.JWTSecret)
		tokenString, err := authConfig.GenerateToken(user.Username, user.Role, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not generate token",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   tokenString,
		})
	}

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		authConfig = auth.NewAuthConfig("super-secret-key")
		"message": "Login successful",
		"token":   tokenString,
	})

		protected := r.Group("/")
		protected.Use(authmw.JWTMiddleware(authConfig))
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out",
	})

func main() {
	connectDB()

	r := gin.Default()

	r.POST("/login", login)

	protected := r.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/me", getCurrentUser)
	}

	log.Println("User service running :8081")
	r.Run(":8081")
}