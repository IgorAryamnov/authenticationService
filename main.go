package main

import (
	"github.com/gin-gonic/gin"
)

type user struct {
    GUID string `json:"guid"`
    Login string `json:"login"`
    Password string `json:"password"`
}
type tokens struct {
    AccessToken string `json:"accessToken"`
    RefreshToken string `json:"refreshToken"`
}
type data struct {
    Login string `bson:"login"`
    RefreshToken string `bson:"refreshToken"`
}
type refreshData struct {
    Refresh string `json:"refresh"`
}
const uri = "mongodb://localhost:27017"
var jwtSecretKey = []byte("very-secret-key")

func main() {
    router := gin.Default()
    router.POST("/registration", CreateNewUser)
    router.POST("/authorization", GetToken)
    router.POST("/refresh", RefreshTokens)
    router.Run("localhost:8080")
}

// curl http://localhost:8080/registration ^
//     --include ^
//     --header "Content-Type: application/json" ^
//     --request "POST" ^
//     --data '{"login": "Name", "password": "Password"}'

// curl http://localhost:8080/authorization ^
//     --include ^
//     --header "Content-Type: application/json" ^
//     --request "POST" ^
//     --data '{"login": "Login", "password": "Password", "guid": "c3d0dc94-1f35-42b7-99cf-54267d6494c0"}'

// curl http://localhost:8080/refresh ^
//     --include ^
//     --header "Content-Type: application/json" ^
//     --header "Authorization: Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTIxNDgwMzcsInN1YiI6IkxvZ2luIn0.0qdVAP7Td4OU_fntc0a9VKu0t3SN4umFY7ajOgK01QZzw1td3O2nFV5yxjrPa5M5Fb6mhpw5mlc8HNeQ1AwkAA" ^
//     --request "POST" ^
//     --data '{"refresh": "99ed0d03f6e16801786040189aa4ce57a9897eda86cf43fdfb0fdfd1509ecab0"}'