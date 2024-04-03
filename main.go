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