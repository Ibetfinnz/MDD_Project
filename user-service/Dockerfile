FROM golang:alpine

RUN apk add --no-cache git

WORKDIR /app

# copy source ทั้งหมดก่อน
COPY . .

COPY ../common /common


# โหลด dependency
RUN go mod download

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

EXPOSE 8081

CMD ["./main"]