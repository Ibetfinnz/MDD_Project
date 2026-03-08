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
	log.Println("User Service: health check")
	c.String(200, "User Service Running")
}

// Get all users
func getAllUsers(c *gin.Context) {
	log.Println("User Service: get all users")
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

	log.Printf("User Service: login attempt for user=%s", input.Username)

	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		log.Printf("User Service: login failed for user=%s (user not found)", input.Username)
		c.JSON(401, gin.H{"error": "Invalid login"})
		return
	}

	if user.Password != input.Password {
		log.Printf("User Service: login failed for user=%s (wrong password)", input.Username)
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

	log.Printf("User Service: login success for user=%s role=%s", user.Username, user.Role)

	c.JSON(200, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

// Get current user info from JWT
func getCurrentUser(c *gin.Context) {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(200, gin.H{
		"username": user.Username,
		"role":     user.Role,
	})
}

func main() {
	connectDB()

	authConfig = auth.NewAuthConfig("super-secret-key")

	r := gin.Default()

	r.GET("/", serviceCheck)
	r.POST("/login", login)

	userGroup := r.Group("/")
	userGroup.Use(middleware.RequireUser())
	{
		userGroup.GET("/me", getCurrentUser)
	}

	adminGroup := r.Group("/")
	adminGroup.Use(middleware.RequireAdmin())
	{
		adminGroup.GET("/users", getAllUsers)
	}

	log.Println("User service running on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
