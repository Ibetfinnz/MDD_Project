package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type WaterMeter struct {
	gorm.Model
	RoomID string  `json:"room_id"`
	Unit   float64 `json:"unit"`
	Month  string  `json:"month"`
}

type ElectricMeter struct {
	gorm.Model
	RoomID string  `json:"room_id"`
	Unit   float64 `json:"unit"`
	Month  string  `json:"month"`
}

var db *gorm.DB

func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("meter.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to meter database: ", err)
	}
	log.Println("✅ Meter Database connected!")
	db.AutoMigrate(&WaterMeter{}, &ElectricMeter{})
}

func main() {
	connectDB()

	r := gin.Default()
	r.Use(cors.Default())

	meterRoute := r.Group("/meter")
	{
		// --- WATER METER ---
		// 🆕 ดูประวัติการจดมิเตอร์น้ำทั้งหมด
		meterRoute.GET("/water", func(c *gin.Context) {
			var meters []WaterMeter
			db.Find(&meters)
			c.JSON(http.StatusOK, meters)
		})

		meterRoute.GET("/water/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var meter WaterMeter
			db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
			c.JSON(http.StatusOK, meter)
		})

		meterRoute.POST("/water", func(c *gin.Context) {
			var meter WaterMeter
			if err := c.ShouldBindJSON(&meter); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}
			meter.Month = time.Now().Format("2006-01")
			db.Create(&meter)
			c.JSON(http.StatusCreated, meter)
		})

		// --- ELECTRIC METER ---
		// 🆕 ดูประวัติการจดมิเตอร์ไฟทั้งหมด
		meterRoute.GET("/electric", func(c *gin.Context) {
			var meters []ElectricMeter
			db.Find(&meters)
			c.JSON(http.StatusOK, meters)
		})

		meterRoute.GET("/electric/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var meter ElectricMeter
			db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
			c.JSON(http.StatusOK, meter)
		})

		meterRoute.POST("/electric", func(c *gin.Context) {
			var meter ElectricMeter
			if err := c.ShouldBindJSON(&meter); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}
			meter.Month = time.Now().Format("2006-01")
			db.Create(&meter)
			c.JSON(http.StatusCreated, meter)
		})
	}

	log.Println("🚀 Meter Service is running on port 8083...")
	r.Run(":8083")
}
