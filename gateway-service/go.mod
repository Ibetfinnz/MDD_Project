module gateway-service

go 1.25.0

require (
	github.com/Ibetfinnz/MDD_Project/auth v0.0.0
	github.com/sony/gobreaker v1.0.0
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace github.com/Ibetfinnz/MDD_Project/auth => ../auth
