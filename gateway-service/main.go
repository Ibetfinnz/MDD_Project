package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ฟังก์ชันสำหรับจัดการ CORS (เปิดประตูให้ Frontend)
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// อนุญาตให้ทุก Domain เข้าถึงได้ (หรือระบุ http://localhost:5173 ก็ได้ครับ)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// ถ้าเป็นคำสั่ง OPTIONS (ที่เบราว์เซอร์ยิงมาเช็คก่อน) ให้ตอบกลับ 200 ทันที
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func setupProxy(target string) http.Handler {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return proxy
}

func main() {
	// สร้าง ServeMux เพื่อจัดการ Route
	mux := http.NewServeMux()

	// ตั้งค่า Proxy โดยชี้ไปยังชื่อ Service ใน Docker
	mux.Handle("/api/users/", http.StripPrefix("/api/users/", setupProxy("http://user-service:8081")))
	mux.Handle("/api/rooms/", http.StripPrefix("/api/rooms/", setupProxy("http://room-service:8082")))
	mux.Handle("/api/meters/", http.StripPrefix("/api/meters/", setupProxy("http://meter-service:8083")))
	mux.Handle("/api/bills/", http.StripPrefix("/api/bills/", setupProxy("http://bill-service:8084")))
	log.Println("🚀 API Gateway with CORS is running on port 8080...")

	// รัน Server โดยครอบด้วย Middleware CORS ที่เราสร้างไว้
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
