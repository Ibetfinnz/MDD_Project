package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// WaterMeter represents water usage per room/month
type WaterMeter struct {
	gorm.Model
	RoomID string  `json:"room_id"`
	Unit   float64 `json:"unit"`
	Month  string  `json:"month"`
}

// ElectricMeter represents electricity usage per room/month
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
		log.Fatal("Failed to connect to meter database: ", err)
	}
	log.Println("Meter database connected")
	db.AutoMigrate(&WaterMeter{}, &ElectricMeter{})
}

// RabbitMQ setup
func connectRabbitMQ() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	for {
		conn, err := amqp.Dial(rabbitURL)
		if err != nil {
			log.Println("Failed to connect to RabbitMQ, retry in 3s:", err)
			time.Sleep(3 * time.Second)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			log.Println("Failed to open channel, retry in 3s:", err)
			conn.Close()
			time.Sleep(3 * time.Second)
			continue
		}

		queues := []string{"meter.water.created", "meter.electric.created"}
		for _, q := range queues {
			_, err = ch.QueueDeclare(
				q,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				log.Println("Failed to declare queue:", q, err)
			}
		}

		rabbitConn = conn
		rabbitCh = ch

		log.Println("Meter service connected to RabbitMQ")
		break
	}
}

func publishEvent(queue string, payload any) {
	if rabbitCh == nil {
		log.Println("RabbitMQ channel not ready, skip publish to", queue)
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal event:", err)
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
		log.Println("Failed to publish event to", queue, ":", err)
	} else {
		log.Println("Published event to", queue)
	}
}

// GET /water
func getAllWaterMeters(c *gin.Context) {
	_, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	log.Println("Meter Service: get all water meters")
	var meters []WaterMeter
	db.Find(&meters)
	c.JSON(200, meters)
}

// GET /water/:room_id
func getWaterMeterByRoom(c *gin.Context) {
	_, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	roomID := c.Param("room_id")
	log.Printf("Meter Service: get water meter for room_id=%s", roomID)
	var meter WaterMeter
	db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
	c.JSON(200, meter)
}

// POST /water
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

	log.Printf("Meter Service: create water meter room_id=%s unit=%.2f by user=%s", meter.RoomID, meter.Unit, user.Username)
	meter.Month = time.Now().Format("2006-01")
	db.Create(&meter)

	publishEvent("meter.water.created", meter)

	c.JSON(201, meter)
}

// GET /electric
func getAllElectricMeters(c *gin.Context) {
	var meters []ElectricMeter
	db.Find(&meters)
	c.JSON(200, meters)
}

// GET /electric/:room_id
func getElectricMeterByRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	var meter ElectricMeter
	db.Where("room_id = ?", roomID).Order("created_at desc").First(&meter)
	c.JSON(200, meter)
}

// POST /electric
func createElectricMeter(c *gin.Context) {
	var meter ElectricMeter
	if err := c.ShouldBindJSON(&meter); err != nil {
		c.JSON(400, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}

	log.Printf("Meter Service: create electric meter room_id=%s unit=%.2f", meter.RoomID, meter.Unit)
	meter.Month = time.Now().Format("2006-01")
	db.Create(&meter)

	publishEvent("meter.electric.created", meter)

	c.JSON(201, meter)
}

func main() {
	connectDB()
	connectRabbitMQ()

	r := gin.Default()

	authorized := r.Group("/")
	authorized.Use(middleware.RequireUser())
	{
		authorized.GET("/water", getAllWaterMeters)
		authorized.GET("/water/:room_id", getWaterMeterByRoom)
		authorized.GET("/electric", getAllElectricMeters)
		authorized.GET("/electric/:room_id", getElectricMeterByRoom)

		// Admin endpoints for creating meter data
		admin := authorized.Group("/")
		admin.Use(middleware.RequireAdmin())
		{
			admin.POST("/water", createWaterMeter)
			admin.POST("/electric", createElectricMeter)
		}
	}

	log.Println("Meter Service is running on port 8083")
	r.Run(":8083")
}
