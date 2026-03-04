package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

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

// --------- ข้อมูลจาก service อื่น ---------

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

// ดึงข้อมูลห้องจาก room-service โดยใช้ RoomNumber = roomID
func fetchRoom(roomID string) (*Room, error) {
	resp, err := http.Get("http://room-service:8082/rooms/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("room-service status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rooms []Room
	if err := json.Unmarshal(body, &rooms); err != nil {
		return nil, err
	}

	for _, r := range rooms {
		if r.RoomNumber == roomID {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("room not found")
}

// ดึงค่าน้ำล่าสุดจาก meter-service
func fetchLatestWater(roomID string) (*WaterMeter, error) {
	url := fmt.Sprintf("http://meter-service:8083/meter/water/%s", roomID)
	resp, err := http.Get(url)
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

// ดึงค่าไฟล่าสุดจาก meter-service
func fetchLatestElectric(roomID string) (*ElectricMeter, error) {
	url := fmt.Sprintf("http://meter-service:8083/meter/electric/%s", roomID)
	resp, err := http.Get(url)
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

func connectDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("bill.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to bill database: ", err)
	}
	log.Println("✅ Bill Database connected!")
	db.AutoMigrate(&Bill{})
}

func main() {
	connectDB()

	r := gin.Default()
	r.Use(cors.Default())

	// 🆕 GET /Bill = ดูบิลทั้งหมดของทุกห้อง
	r.GET("/Bill", func(c *gin.Context) {
			var bills []Bill
			db.Find(&bills)
			c.JSON(http.StatusOK, bills)
		})

	// GET /Bill/:room_id = ดูบิลล่าสุดของห้องนั้น
	r.GET("/Bill/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var bill Bill
			if err := db.Where("room_id = ?", roomID).Order("created_at desc").First(&bill).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลบิลสำหรับห้องนี้"})
				return
			}
			c.JSON(http.StatusOK, bill)
		})

	// POST /Bill/:room_id = สร้างบิลค่าเช่าใหม่
	r.POST("/Bill/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var bill Bill
			if err := c.ShouldBindJSON(&bill); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}

			// ผูก room_id ให้ชัดเจน
			bill.RoomID = roomID
			bill.Month = time.Now().Format("2006-01")

			// --- ดึงค่าเช่าจาก room-service ---
			if room, err := fetchRoom(roomID); err == nil {
				bill.RentPrice = room.Price
			} else {
				log.Println("⚠️ fetchRoom error:", err)
			}

			// --- ดึงค่าน้ำ/ค่าไฟจาก meter-service แล้วคำนวณราคา ---
			if water, err := fetchLatestWater(roomID); err == nil {
				bill.WaterPrice = water.Unit * waterRatePerUnit
			} else {
				log.Println("⚠️ fetchLatestWater error:", err)
			}

			if electric, err := fetchLatestElectric(roomID); err == nil {
				bill.ElectricPrice = electric.Unit * electricRatePerUnit
			} else {
				log.Println("⚠️ fetchLatestElectric error:", err)
			}

			// รวมยอด
			bill.Total = bill.RentPrice + bill.WaterPrice + bill.ElectricPrice
			if bill.Status == "" {
				bill.Status = "Unpaid"
			}

			if err := db.Create(&bill).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้างบิลไม่สำเร็จ"})
				return
			}
			c.JSON(http.StatusCreated, bill)
		})

	// PATCH /Bill/:room_id = แก้ไขสถานะการจ่ายเงิน หรือยอดเงิน
	r.PATCH("/Bill/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var bill Bill
			if err := db.Where("room_id = ?", roomID).Order("created_at desc").First(&bill).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลบิล"})
				return
			}

			c.ShouldBindJSON(&bill)
			bill.Total = bill.RentPrice + bill.WaterPrice + bill.ElectricPrice
			db.Save(&bill)

			c.JSON(http.StatusOK, gin.H{"message": "แก้ไขบิลค่าเช่าสำเร็จ", "data": bill})
		})

	log.Println("🚀 Bill Service is running on port 8084...")
	r.Run(":8084")
}
