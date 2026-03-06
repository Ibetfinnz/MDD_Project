FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

# copy source ทั้งหมดก่อน
COPY . .

# โหลด dependency
RUN go mod download

# เปิดพอร์ต (ให้แก้เลข 8081 นี้ให้ตรงกับพอร์ตของแต่ละ Service นะครับ)
EXPOSE 8083

# สั่งรันไฟล์ที่ Build เสร็จแล้ว
CMD ["./main"]