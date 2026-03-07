package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
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
var rabbitConn *amqp.Connection
var rabbitCh *amqp.Channel

func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("meter.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to meter database: ", err)
	}
	log.Println("✅ Meter Database connected!")
	db.AutoMigrate(&WaterMeter{}, &ElectricMeter{})
}

// --- RabbitMQ Setup ---
func connectRabbitMQ() {
	var err error
	rabbitConn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Println("⚠️ Failed to connect to RabbitMQ:", err)
		return
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		log.Println("⚠️ Failed to open channel:", err)
		return
	}

	// Simple durable queues for water & electric events
	queues := []string{"meter.water.created", "meter.electric.created"}
	for _, q := range queues {
		_, err = rabbitCh.QueueDeclare(
			q,
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args
		)
		if err != nil {
			log.Println("⚠️ Failed to declare queue:", q, err)
		}
	}

	log.Println("✅ Meter Service connected to RabbitMQ")
}

func publishEvent(queue string, payload any) {
	if rabbitCh == nil {
		// ถ้า RabbitMQ ใช้งานไม่ได้ ก็แค่ log แล้วข้าม (ไม่ให้ล้ม service)
		log.Println("⚠️ RabbitMQ channel not ready, skip publish to", queue)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("⚠️ Failed to marshal event:", err)
		return
	}

	err = rabbitCh.Publish(
		"",    // default exchange
		queue, // routing key = queue name
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		log.Println("⚠️ Failed to publish event to", queue, ":", err)
	} else {
		log.Println("📨 Published event to", queue)
	}
}

// Handler: GET /water - ดูประวัติการจดมิเตอร์น้ำทั้งหมด
func getAllWaterMeters(c *gin.Context) {
	_, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	var meters []WaterMeter
	db.Find(&meters)
	c.JSON(200, meters)
}

// Handler: GET /water/:room_id - ดูมิเตอร์น้ำล่าสุดของห้อง
func getWaterMeterByRoom(c *gin.Context) {
	_, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	roomID := c.Param("room_id")
	var meter WaterMeter
	db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
	c.JSON(200, meter)
}

// Handler: POST /water - บันทึกมิเตอร์น้ำใหม่
func createWaterMeter(c *gin.Context) {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	if user.Role != "admin" {
		c.JSON(403, gin.H{"error": "เฉพาะ admin เท่านั้น"})
		return
	}

	var meter WaterMeter
	if err := c.ShouldBindJSON(&meter); err != nil {
		c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	meter.Month = time.Now().Format("2006-01")
	db.Create(&meter)

	// publish event ไป RabbitMQ (async)
	publishEvent("meter.water.created", meter)

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

	// publish event ไป RabbitMQ (async)
	publishEvent("meter.electric.created", meter)

	c.JSON(201, meter)
}

func main() {
	connectDB()
	connectRabbitMQ()

	r := gin.Default()
	r.Use(cors.Default(), middleware.RequireUser())

	// --- WATER METER ---
	r.GET("/water", getAllWaterMeters)
	r.GET("/water/:room_id", getWaterMeterByRoom)

	admin := r.Group("/")
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/water", createWaterMeter)
	}

	// --- ELECTRIC METER ---
	r.GET("/electric", getAllElectricMeters)
	r.GET("/electric/:room_id", getElectricMeterByRoom)
	admin.POST("/electric", createElectricMeter)

	log.Println("🚀 Meter Service is running on port 8083...")
	r.Run(":8083")
}
