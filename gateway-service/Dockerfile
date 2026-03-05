# ใช้ Go เวอร์ชั่นล่าสุดแบบเบาบาง (Alpine)
FROM golang:alpine

# ติดตั้ง git สำหรับ go mod
RUN apk add --no-cache git

# กำหนดโฟลเดอร์ทำงานข้างใน Docker
WORKDIR /app

# ก๊อปปี้ไฟล์จัดการแพ็กเกจเข้ามาก่อน

# โหลดแพ็กเกจที่จำเป็น

# ก๊อปปี้ซอร์สโค้ดทั้งหมดเข้าไป
COPY . .

# สั่ง Build โค้ด Go ให้เป็นไฟล์รันได้ (ใช้ CGO_ENABLED=0 เพื่อไม่ให้มีปัญหาเหมือนตอนแรกรันบน Windows)
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# เปิดพอร์ต (ให้แก้เลข 8081 นี้ให้ตรงกับพอร์ตของแต่ละ Service นะครับ)
EXPOSE 8080

# สั่งรันไฟล์ที่ Build เสร็จแล้ว
CMD ["./main"]