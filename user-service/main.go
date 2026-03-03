package main

import (
	"log"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

var db *gorm.DB
var currentRole string // ✅ เก็บ role ล่าสุด

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

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// --- Service Check ---
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("User Service Running")
	})

	// --- Get All Users ---
	app.Get("/users", func(c *fiber.Ctx) error {
		var users []User
		if err := db.Select("username, role").Find(&users).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
		}
		return c.JSON(users)
	})

	// ===== LOGIN =====
	app.Post("/login", func(c *fiber.Ctx) error {
		var input User
		if err := c.BodyParser(&input); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid format"})
		}

		var user User
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		if user.Password != input.Password {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		// ✅ เก็บ role ล่าสุด
		currentRole = user.Role

		return c.JSON(fiber.Map{
			"message": "Login successful",
			"role":    currentRole,
		})
	})

	// ===== LOGOUT =====
	app.Post("/logout", func(c *fiber.Ctx) error {
		currentRole = ""
		return c.JSON(fiber.Map{"message": "Logged out"})
	})

	// ===== CHECK ROLE (Service อื่นเรียกมา) =====
	app.Get("/check-role", func(c *fiber.Ctx) error {
		if currentRole == "" {
			return c.Status(401).JSON(fiber.Map{"error": "No user logged in"})
		}

		return c.JSON(fiber.Map{
			"role": currentRole,
		})
	})

	log.Println("Running on :8081")
	log.Fatal(app.Listen(":8081"))
}