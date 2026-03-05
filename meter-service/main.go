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

// Handler: GET /water - ดูประวัติการจดมิเตอร์น้ำทั้งหมด
func getAllWaterMeters(c *gin.Context) {
	var meters []WaterMeter
	db.Find(&meters)
	c.JSON(200, meters)
}

// Handler: GET /water/:room_id - ดูมิเตอร์น้ำล่าสุดของห้อง
func getWaterMeterByRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	var meter WaterMeter
	db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
	c.JSON(200, meter)
}

// Handler: POST /water - บันทึกมิเตอร์น้ำใหม่
func createWaterMeter(c *gin.Context) {
	var meter WaterMeter
	if err := c.ShouldBindJSON(&meter); err != nil {
		c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	meter.Month = time.Now().Format("2006-01")
	db.Create(&meter)
	c.JSON(201, meter)
}

// Handler: GET /electric - ดูประวัติการจดมิเตอร์ไฟทั้งหมด
func getAllElectricMeters(c *gin.Context) {
	var meters []ElectricMeter
	db.Find(&meters)
	c.JSON(200, meters)
}

// Handler: GET /electric/:room_id - ดูมิเตอร์ไฟล่าสุดของห้อง
func getElectricMeterByRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	var meter ElectricMeter
	db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
	c.JSON(200, meter)
}

// Handler: POST /electric - บันทึกมิเตอร์ไฟใหม่
func createElectricMeter(c *gin.Context) {
	var meter ElectricMeter
	if err := c.ShouldBindJSON(&meter); err != nil {
		c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	meter.Month = time.Now().Format("2006-01")
	db.Create(&meter)
	c.JSON(201, meter)
}

func main() {
	connectDB()

	r := gin.Default()
	r.Use(cors.Default())

	// --- WATER METER ---
	r.GET("/water", getAllWaterMeters)
	r.GET("/water/:room_id", getWaterMeterByRoom)
	r.POST("/water", createWaterMeter)

	// --- ELECTRIC METER ---
	r.GET("/electric", getAllElectricMeters)
	r.GET("/electric/:room_id", getElectricMeterByRoom)
	r.POST("/electric", createElectricMeter)

	log.Println("🚀 Meter Service is running on port 8083...")
	r.Run(":8083")
}
