package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/sony/gobreaker"
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

// breakerTransport implements http.RoundTripper and wraps requests with a circuit breaker.
type breakerTransport struct {
	cb *gobreaker.CircuitBreaker
	rt http.RoundTripper
}

func (b *breakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	result, err := b.cb.Execute(func() (interface{}, error) {
		return b.rt.RoundTrip(req)
	})
	if err != nil {
		return nil, err
	}
	resp, _ := result.(*http.Response)
	return resp, nil
}

func newCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5, // in HALF-OPEN state
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
	}
	return gobreaker.NewCircuitBreaker(settings)
}

func setupProxy(target, name string) http.Handler {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)

	cb := newCircuitBreaker(name)
	transport := &breakerTransport{
		cb: cb,
		rt: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	proxy.Transport = transport

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// If the breaker is open, fail fast with 503
		if err == gobreaker.ErrOpenState {
			http.Error(w, "service temporarily unavailable (circuit breaker open)", http.StatusServiceUnavailable)
			return
		}
		// default behavior: 502
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return proxy
}

func main() {
	// สร้าง ServeMux เพื่อจัดการ Route
	mux := http.NewServeMux()

	// ตั้งค่า Proxy โดยชี้ไปยังชื่อ Service ใน Docker
	mux.Handle("/api/users/", http.StripPrefix("/api/users/", setupProxy("http://user-service:8081", "user-service")))
	mux.Handle("/api/rooms/", http.StripPrefix("/api/rooms/", setupProxy("http://room-service:8082", "room-service")))
	mux.Handle("/api/meters/", http.StripPrefix("/api/meters/", setupProxy("http://meter-service:8083", "meter-service")))
	mux.Handle("/api/bills/", http.StripPrefix("/api/bills/", setupProxy("http://bill-service:8084", "bill-service")))
	log.Println("🚀 API Gateway with CORS is running on port 8080...")

	// รัน Server โดยครอบด้วย Middleware CORS ที่เราสร้างไว้
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
