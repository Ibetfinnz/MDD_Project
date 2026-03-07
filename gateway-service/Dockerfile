FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/gateway-service

RUN go mod download

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]