package main

import (
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
	Type       string  `json:"type"` // เช่น Standard, VIP
	Price      float64 `json:"price"`
	Status     string  `json:"status"` // เช่น Available, Occupied
	TenantName string  `json:"tenant_name"`
}

var db *gorm.DB

// 2. ฟังก์ชันเชื่อมต่อฐานข้อมูล
func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("room.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database: ", err)
	}
	log.Println("✅ Room Database connected!")
	db.AutoMigrate(&Room{}) // สร้างตารางอัตโนมัติ
}

func main() {
	connectDB()

	r := gin.Default()

	// 3. ตั้งค่า CORS (สำคัญเพื่อให้เชื่อมต่อกับ Gateway/Frontend ได้)
	r.Use(cors.Default())

	roomRoute := r.Group("/rooms")
	{
		// ดูห้องทั้งหมด
		roomRoute.GET("/", func(c *gin.Context) {
			var rooms []Room
			db.Find(&rooms)
			c.JSON(http.StatusOK, rooms)
		})

		// ดูรายละเอียดห้องตาม ID
		roomRoute.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var room Room
			if err := db.First(&room, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องพักที่ระบุ"})
				return
			}
			c.JSON(http.StatusOK, room)
		})

		// เพิ่มห้องพักใหม่ (POST /rooms)
		roomRoute.POST("/", func(c *gin.Context) {
			var room Room
			if err := c.ShouldBindJSON(&room); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}
			db.Create(&room)
			c.JSON(http.StatusCreated, room)
		})

		// แก้ไขรายละเอียดห้อง
		roomRoute.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var room Room
			if err := db.First(&room, id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบห้องพัก"})
				return
			}
			c.ShouldBindJSON(&room)
			db.Save(&room)
			c.JSON(http.StatusOK, gin.H{"message": "แก้ไขรายละเอียดห้องสำเร็จ", "data": room})
		})

		// เพิ่มผู้เช่าเข้าห้อง (อัปเดตชื่อผู้เช่าและสถานะ)
		roomRoute.POST("/:id/tenant", func(c *gin.Context) {
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

			c.JSON(http.StatusOK, gin.H{"message": "เพิ่มผู้เช่าเข้าห้องสำเร็จ", "data": room})
		})
	}

	log.Println("🚀 Room Service is running on port 8082...")
	r.Run(":8082")
}
