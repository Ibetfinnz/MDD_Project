FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

# copy source ทั้งหมดก่อน (ใช้ context ทั้งโปรเจกต์จาก docker-compose)
COPY . .


# โหลด dependency
RUN go mod download

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8081

CMD ["./main"]