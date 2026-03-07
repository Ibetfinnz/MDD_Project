FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

COPY . .

WORKDIR /app/user-service

RUN go mod download

RUN go build -o main .

EXPOSE 8081

CMD ["./main"]