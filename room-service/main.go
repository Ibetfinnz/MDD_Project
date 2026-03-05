package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
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

type CurrentUser struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

var db *gorm.DB

func getCurrentUser() (*CurrentUser, error) {
	resp, err := http.Get("http://user-service:8081/check-role")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("not authorized")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user CurrentUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Handler: GET /rooms
func getRooms(c *gin.Context) {
	_, err := getCurrentUser()
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	var rooms []Room
	db.Find(&rooms)
	c.JSON(200, rooms)
}

// Handler: GET /rooms/:id
func getRoomByID(c *gin.Context) {
	_, err := getCurrentUser()
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	id := c.Param("id")
	var room Room
	if err := db.First(&room, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบห้องพัก"})
		return
	}

	c.JSON(200, room)
}

// Handler: POST /rooms
func createRoom(c *gin.Context) {
	user, err := getCurrentUser()
	if err != nil || user.Role != "admin" {
		c.JSON(403, gin.H{"error": "เฉพาะ admin เท่านั้น"})
		return
	}

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
	user, err := getCurrentUser()
	if err != nil || user.Role != "admin" {
		c.JSON(403, gin.H{"error": "เฉพาะ admin เท่านั้น"})
		return
	}

	id := c.Param("id")
	var room Room
	if err := db.First(&room, id).Error; err != nil {
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
	user, err := getCurrentUser()
	if err != nil || user.Role != "admin" {
		c.JSON(403, gin.H{"error": "เฉพาะ admin เท่านั้น"})
		return
	}

	id := c.Param("id")
	var room Room
	if err := db.First(&room, id).Error; err != nil {
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

	// 3. ตั้งค่า CORS (สำคัญเพื่อให้เชื่อมต่อกับ Gateway/Frontend ได้)
	r.Use(cors.Default())

	r.GET("/rooms", getRooms)
	r.GET("/rooms/:id", getRoomByID)
	r.POST("/rooms", createRoom)
	r.PATCH("/rooms/:id", updateRoom)
	r.POST("/rooms/:id/tenant", addTenantToRoom)

	log.Println("🚀 Room Service is running on port 8082...")
	r.Run(":8082")
}
