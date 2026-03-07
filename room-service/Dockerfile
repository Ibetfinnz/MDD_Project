FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/room-service

RUN go mod download

RUN go build -o main .

EXPOSE 8082

CMD ["./main"]