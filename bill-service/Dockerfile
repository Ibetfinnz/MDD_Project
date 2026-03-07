FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/bill-service

RUN go mod download

RUN go build -o main .

EXPOSE 8084

CMD ["./main"]