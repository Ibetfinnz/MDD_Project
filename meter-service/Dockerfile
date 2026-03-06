FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/meter-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8083

CMD ["./main"]