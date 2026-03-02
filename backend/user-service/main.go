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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

var db *gorm.DB

// 2. ฟังก์ชันเชื่อมต่อฐานข้อมูล
func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("user.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database: \n", err)
	}
	log.Println("✅ Database connected successfully!")

	db.AutoMigrate(&User{})
	log.Println("✅ Database Migrated!")
}

// 3. ฟังก์ชันหลัก (Main)
func main() {
	connectDB()

	app := fiber.New()

	// เปิดใช้งาน CORS ให้ Frontend (React) เรียกใช้ API ได้
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// --- Route 1: เช็คสถานะ ---
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🟢 User Service is up and running!")
	})

	// --- Route 2: ดึงข้อมูล User ทั้งหมด ---
	app.Get("/users", func(c *fiber.Ctx) error {
		var users []User
		db.Find(&users)
		return c.JSON(users)
	})

	// --- Route 3: สร้าง User ใหม่ (ไม่เข้ารหัสผ่าน) ---
	app.Post("/users", func(c *fiber.Ctx) error {
		user := new(User)

		// 1. รับข้อมูล JSON จาก Request
		if err := c.BodyParser(user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
		}

		// 2. ป้องกันไม่ให้ส่งรหัสผ่านว่างๆ มา
		if user.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุรหัสผ่าน"})
		}

		// 3. บันทึกลง Database ทันที
		db.Create(&user)

		return c.Status(201).JSON(user)
	})

	log.Println("🚀 Starting User Service on port 8081...")
	log.Fatal(app.Listen(":8081"))
}
