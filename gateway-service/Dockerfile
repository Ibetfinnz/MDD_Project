FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/gateway-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8080

CMD ["./main"]