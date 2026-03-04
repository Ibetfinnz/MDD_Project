package main

import (
	"log"
	"net/http"

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

var db *gorm.DB
var currentUser *User

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

func main() {
	connectDB()

	r := gin.Default()

	// --- Service Check ---
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "User Service Running")
	})

	// --- Get All Users ---
	r.GET("/users", func(c *gin.Context) {
		var users []User
		if err := db.Select("username, role").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch users",
			})
			return
		}
		c.JSON(http.StatusOK, users)
	})

	// ===== LOGIN =====
	r.POST("/login", func(c *gin.Context) {
		var input User
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid format",
			})
			return
		}

		var user User
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
			return
		}

		if user.Password != input.Password {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
			return
		}

		currentUser = &user

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"username"	currentUser.Username,
			"role":    currentUser.Role,
		})
	})

	// ===== LOGOUT =====
	r.POST("/logout", func(c *gin.Context) {
		currentUser = nil
		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out",
		})
	})

	// ===== CHECK ROLE =====
	r.GET("/check-role", func(c *gin.Context) {
		if currentUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No user logged in",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"username": currentUser.Username,
			"role": currentUser.Role,
		})
	})

	log.Println("Running on :8081")
	r.Run(":8081")
}