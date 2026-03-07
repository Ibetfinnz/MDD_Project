package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	"gorm.io/gorm"
)

// 1. โครงสร้างข้อมูล Room (Data Model)
type Room struct {
	gorm.Model
	RoomNumber string  `json:"room_number" gorm:"unique"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`      // Available / Occupied
	TenantName string  `json:"tenant_name"` // ว่างได้ถ้ายังไม่มีผู้เช่า
}

var db *gorm.DB

// Handler: GET /rooms
func getRooms(c *gin.Context) {
	var rooms []Room
	db.Find(&rooms)
	c.JSON(200, rooms)
}

// Handler: GET /rooms/:id
func getRoomByID(c *gin.Context) {
	roomNumber := c.Param("id")
	var room Room
	if err := db.Where("room_number = ?", roomNumber).First(&room).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบห้องพัก"})
		return
	}

	c.JSON(200, room)
}

// Handler: POST /rooms
func createRoom(c *gin.Context) {
	var room Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	db.Create(&room)
	c.JSON(201, room)
}

// Handler: PATCH /rooms/:id
func updateRoom(c *gin.Context) {
	roomNumber := c.Param("id")
	var room Room
	if err := db.Where("room_number = ?", roomNumber).First(&room).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบห้องพัก"})
		return
	}

	c.ShouldBindJSON(&room)
	db.Save(&room)

	c.JSON(200, gin.H{
		"message": "แก้ไขสำเร็จ",
		"data":    room,
	})
}

// Handler: POST /rooms/:id/tenant
func addTenantToRoom(c *gin.Context) {
	roomNumber := c.Param("id")
	var room Room
	if err := db.Where("room_number = ?", roomNumber).First(&room).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบห้องพัก"})
		return
	}

	var input struct {
		TenantName string `json:"tenant_name"`
	}

	c.ShouldBindJSON(&input)

	room.TenantName = input.TenantName
	room.Status = "Occupied"
	db.Save(&room)

	c.JSON(200, gin.H{
		"message": "เพิ่มผู้เช่าสำเร็จ",
		"data":    room,
	})
}

// 2. ฟังก์ชันเชื่อมต่อฐานข้อมูล
func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("room.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database: ", err)
	}

	log.Println("✅ Room Database connected!")

	// AutoMigrate โครงสร้างใหม่
	err = db.AutoMigrate(&Room{})
	if err != nil {
		log.Fatal("❌ Migration failed: ", err)
	}

	log.Println("✅ AutoMigrate completed!")

	// Seed ถ้ายังไม่มีข้อมูล
	var count int64
	db.Model(&Room{}).Count(&count)

	if count == 0 {
		log.Println("🌱 Seeding initial room data...")

		rooms := []Room{
			{RoomNumber: "101", Price: 3500, Status: "Occupied", TenantName: "tenant1"},
			{RoomNumber: "102", Price: 3500, Status: "Available"},
			{RoomNumber: "201", Price: 6000, Status: "Available"},
		}

		for _, room := range rooms {
			db.Create(&room)
		}

		log.Println("✅ Seed data inserted!")
	}
}

func main() {
	connectDB()

	r := gin.Default()

	// กลุ่มสำหรับทุก endpoint ที่ต้อง login แล้ว
	authorized := r.Group("/")
	authorized.Use(middleware.RequireUser())
	{
		// สิทธิ์ของ user ทั่วไป (เช่น tenant) ดูรายการห้อง/รายละเอียดห้องได้
		authorized.GET("/", getRooms)
		authorized.GET("/:id", getRoomByID)

		// กลุ่มที่ต้องเป็น admin เท่านั้น
		admin := authorized.Group("/")
		admin.Use(middleware.RequireAdmin())
		{
			admin.POST("/", createRoom)
			admin.PATCH("/:id", updateRoom)
			admin.POST("/:id/tenant", addTenantToRoom)
		}
	}

	log.Println("🚀 Room Service is running on port 8082...")
	r.Run(":8082")
}
