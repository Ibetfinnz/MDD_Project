FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/meter-service

RUN go mod download

RUN go build -o main .

EXPOSE 8083

CMD ["./main"]