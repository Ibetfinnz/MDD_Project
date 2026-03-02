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

	billRoute := r.Group("/Bill")
	{
		// 🆕 GET /Bill/ = ดูบิลทั้งหมดของทุกห้อง (เพิ่มอันนี้ครับ)
		billRoute.GET("/", func(c *gin.Context) {
			var bills []Bill
			db.Find(&bills)
			c.JSON(http.StatusOK, bills)
		})

		// GET /Bill/:room_id = ดูบิลล่าสุดของห้องนั้น
		billRoute.GET("/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var bill Bill
			if err := db.Where("room_id = ?", roomID).Order("created_at desc").First(&bill).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อมูลบิลสำหรับห้องนี้"})
				return
			}
			c.JSON(http.StatusOK, bill)
		})

		// POST /Bill/:room_id = สร้างบิลค่าเช่าใหม่
		billRoute.POST("/:room_id", func(c *gin.Context) {
			roomID := c.Param("room_id")
			var bill Bill
			if err := c.ShouldBindJSON(&bill); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
				return
			}

			bill.RoomID = roomID
			bill.Month = time.Now().Format("2006-01")
			bill.Total = bill.RentPrice + bill.WaterPrice + bill.ElectricPrice
			if bill.Status == "" {
				bill.Status = "Unpaid"
			}

			db.Create(&bill)
			c.JSON(http.StatusCreated, bill)
		})

		// PATCH /Bill/:room_id = แก้ไขสถานะการจ่ายเงิน หรือยอดเงิน
		billRoute.PATCH("/:room_id", func(c *gin.Context) {
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
	}

	log.Println("🚀 Bill Service is running on port 8084...")
	r.Run(":8084")
}
