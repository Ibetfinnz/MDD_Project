package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ibetfinnz/MDD_Project/auth/middleware"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// Bill represents rent and utility charges per room/month
type Bill struct {
	gorm.Model
	RoomID        string  `json:"room_id"`
	RentPrice     float64 `json:"rent_price"`
	WaterPrice    float64 `json:"water_price"`
	ElectricPrice float64 `json:"electric_price"`
	Total         float64 `json:"total"`
	Month         string  `json:"month"`
	Status        string  `json:"status"`
}

// Data from other services
type Room struct {
	RoomNumber string  `json:"room_number"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	TenantName string  `json:"tenant_name"`
}

type WaterMeter struct {
	RoomID string  `json:"room_id"`
	Unit   float64 `json:"unit"`
	Month  string  `json:"month"`
}

type ElectricMeter struct {
	RoomID string  `json:"room_id"`
	Unit   float64 `json:"unit"`
	Month  string  `json:"month"`
}

const (
	waterRatePerUnit    = 10.0 // ปรับเรทค่าน้ำต่อหน่วยได้ตรงนี้
	electricRatePerUnit = 5.0  // ปรับเรทค่าไฟต่อหน่วยได้ตรงนี้
)

// fetchRoom gets room info from room-service and forwards user headers
func fetchRoom(c *gin.Context, roomID string) (*Room, error) {
	url := fmt.Sprintf("http://room-service:8082/%s", roomID)
	log.Printf("Bill Service: call room-service for room_id=%s", roomID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	middleware.AttachUserHeaders(c, req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("room-service status: %d", resp.StatusCode)
	}

	var room Room
	if err := json.NewDecoder(resp.Body).Decode(&room); err != nil {
		return nil, err
	}

	return &room, nil
}

// fetchLatestWater gets latest water meter from meter-service
func fetchLatestWater(c *gin.Context, roomID string) (*WaterMeter, error) {
	url := fmt.Sprintf("http://meter-service:8083/water/%s", roomID)
	log.Printf("Bill Service: call meter-service for latest water room_id=%s", roomID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	middleware.AttachUserHeaders(c, req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("water meter status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m WaterMeter
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

// fetchLatestElectric gets latest electric meter from meter-service
func fetchLatestElectric(c *gin.Context, roomID string) (*ElectricMeter, error) {
	url := fmt.Sprintf("http://meter-service:8083/electric/%s", roomID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	middleware.AttachUserHeaders(c, req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("electric meter status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m ElectricMeter
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

var db *gorm.DB
var rabbitConn *amqp.Connection
var rabbitCh *amqp.Channel

func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("bill.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to bill database: ", err)
	}
	log.Println("Bill database connected")
	db.AutoMigrate(&Bill{})
}

// RabbitMQ consumer
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

		log.Println("Bill service connected to RabbitMQ")

		go consumeMeterEvents("meter.water.created")
		go consumeMeterEvents("meter.electric.created")

		break
	}
}

func consumeMeterEvents(queue string) {
	if rabbitCh == nil {
		return
	}

	msgs, err := rabbitCh.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to register consumer for", queue, ":", err)
		return
	}

	for msg := range msgs {
		log.Printf("Bill Service received from %s: %s\n", queue, string(msg.Body))
	}
}

// GET /Bill
func getAllBills(c *gin.Context) {
	var bills []Bill
	db.Find(&bills)
	c.JSON(200, bills)
}

// GET /Bill/:room_id
func getLatestBillByRoom(c *gin.Context) {
	roomID := c.Param("room_id")
	log.Printf("Bill Service: get latest bill for room_id=%s", roomID)

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "กรุณา login ก่อน"})
		return
	}

	if user.Role != "admin" {
		room, err := fetchRoom(c, roomID)
		if err != nil {
			c.JSON(404, gin.H{"error": "ไม่พบข้อมูลห้องสำหรับบิล"})
			return
		}

		if room.TenantName != user.Username {
			c.JSON(403, gin.H{"error": "ไม่มีสิทธิ์ดูบิลของห้องนี้"})
			return
		}
	}

	var bill Bill
	if err := db.Where("room_id = ?", roomID).Order("created_at desc").First(&bill).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบข้อมูลบิลสำหรับห้องนี้"})
		return
	}
	c.JSON(200, bill)
}

// POST /Bill/:room_id
func createBill(c *gin.Context) {
	roomID := c.Param("room_id")
	var bill Bill

	// ผูก room_id ให้ชัดเจน
	bill.RoomID = roomID
	bill.Month = time.Now().Format("2006-01")

	log.Printf("Bill Service: create bill for room_id=%s", roomID)

	if room, err := fetchRoom(c, roomID); err == nil {
		bill.RentPrice = room.Price
	} else {
		log.Println("⚠️ fetchRoom error:", err)
	}

	if water, err := fetchLatestWater(c, roomID); err == nil {
		bill.WaterPrice = water.Unit * waterRatePerUnit
	} else {
		log.Println("⚠️ fetchLatestWater error:", err)
	}

	if electric, err := fetchLatestElectric(c, roomID); err == nil {
		bill.ElectricPrice = electric.Unit * electricRatePerUnit
	} else {
		log.Println("⚠️ fetchLatestElectric error:", err)
	}

	bill.Total = bill.RentPrice + bill.WaterPrice + bill.ElectricPrice
	if bill.Status == "" {
		bill.Status = "Unpaid"
	}

	log.Printf("Bill Service: calculated bill for room_id=%s rent=%.2f water=%.2f electric=%.2f total=%.2f", bill.RoomID, bill.RentPrice, bill.WaterPrice, bill.ElectricPrice, bill.Total)

	if err := db.Create(&bill).Error; err != nil {
		c.JSON(500, gin.H{"error": "สร้างบิลไม่สำเร็จ"})
		return
	}
	c.JSON(201, bill)
}

// PATCH /Bill/:room_id
func updateBill(c *gin.Context) {
	roomID := c.Param("room_id")
	var bill Bill
	if err := db.Where("room_id = ?", roomID).Order("created_at desc").First(&bill).Error; err != nil {
		c.JSON(404, gin.H{"error": "ไม่พบข้อมูลบิล"})
		return
	}

	c.ShouldBindJSON(&bill)
	bill.Total = bill.RentPrice + bill.WaterPrice + bill.ElectricPrice
	db.Save(&bill)

	c.JSON(200, gin.H{"message": "แก้ไขบิลค่าเช่าสำเร็จ", "data": bill})
}

func main() {
	connectDB()
	connectRabbitMQ()

	r := gin.Default()

	authorized := r.Group("/")
	authorized.Use(middleware.RequireUser())
	{
		// Tenant endpoints
		authorized.GET("/:room_id", getLatestBillByRoom)

		// Admin endpoints
		admin := authorized.Group("/")
		admin.Use(middleware.RequireAdmin())
		{
			admin.GET("/", getAllBills)
			admin.POST("/:room_id", createBill)
			admin.PATCH("/:room_id", updateBill)
		}
	}

	log.Println("Bill Service is running on port 8084")
	r.Run(":8084")
}
