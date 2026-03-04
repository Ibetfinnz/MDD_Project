package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-contrib/cors" // ต้องรัน go get github.com/gin-contrib/cors ด้วยครับ
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

	if resp.StatusCode != http.StatusOK {
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

	roomRoute := r.Group("/rooms")
{
	roomRoute.GET("/", func(c *gin.Context) {

		_, err := getCurrentUser()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
			return
		}

		var rooms []Room
		db.Find(&rooms)
		c.JSON(http.StatusOK, rooms)
	})

	roomRoute.GET("/:id", func(c *gin.Context) {

		_, err := getCurrentUser()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "กรุณา login ก่อน"})
			return
		}

		id := c.Param("id")
		var room Room
		if err := db.First(&room, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องพัก"})
			return
		}

		c.JSON(http.StatusOK, room)
	})

	roomRoute.POST("/", func(c *gin.Context) {

		user, err := getCurrentUser()
		if err != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะ admin เท่านั้น"})
			return
		}

		var room Room
		if err := c.ShouldBindJSON(&room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
			return
		}

		db.Create(&room)
		c.JSON(http.StatusCreated, room)
	})

	roomRoute.PATCH("/:id", func(c *gin.Context) {

		user, err := getCurrentUser()
		if err != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะ admin เท่านั้น"})
			return
		}

		id := c.Param("id")
		var room Room
		if err := db.First(&room, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องพัก"})
			return
		}

		c.ShouldBindJSON(&room)
		db.Save(&room)

		c.JSON(http.StatusOK, gin.H{
			"message": "แก้ไขสำเร็จ",
			"data":    room,
		})
	})

	roomRoute.POST("/:id/tenant", func(c *gin.Context) {

		user, err := getCurrentUser()
		if err != nil || user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "เฉพาะ admin เท่านั้น"})
			return
		}

		id := c.Param("id")
		var room Room
		if err := db.First(&room, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องพัก"})
			return
		}

		var input struct {
			TenantName string `json:"tenant_name"`
		}

		c.ShouldBindJSON(&input)

		room.TenantName = input.TenantName
		room.Status = "Occupied"
		db.Save(&room)

		c.JSON(http.StatusOK, gin.H{
			"message": "เพิ่มผู้เช่าสำเร็จ",
			"data":    room,
		})
	})
}

	log.Println("🚀 Room Service is running on port 8082...")
	r.Run(":8082")
}
