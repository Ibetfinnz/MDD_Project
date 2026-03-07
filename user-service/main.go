package main

import (
	"log"
	"time"

	"github.com/Ibetfinnz/MDD_Project/auth"
	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
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
	c.String(200, "User Service Running")
}

// Get all users
func getAllUsers(c *gin.Context) {
	var users []User
	if err := db.Select("username, role").Find(&users).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}
	c.JSON(200, users)
}

// Login
func login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "Invalid login"})
		return
	}

	if user.Password != input.Password {
		c.JSON(401, gin.H{"error": "Invalid login"})
		return
	}

	tokenString, err := authConfig.GenerateToken(user.Username, user.Role, 24*time.Hour)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not generate token",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

// Get current user info from JWT
func getCurrentUser(c *gin.Context) {
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(200, gin.H{
		"username": username,
		"role":     role,
	})
}

func main() {
	connectDB()

	authConfig = auth.NewAuthConfig("super-secret-key")

	r := gin.Default()

	r.GET("/", serviceCheck)
	r.POST("/login", login)

	authorized := r.Group("/")
	authorized.Use(middleware.JWTMiddleware(authConfig))
	{
		authorized.GET("/me", getCurrentUser)
		authorized.GET("/users", getAllUsers)
	}

	log.Println("User service running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
