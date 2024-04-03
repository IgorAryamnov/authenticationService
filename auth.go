package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetToken(c *gin.Context) {
	var newTokens tokens
	var newUser user

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := c.BindJSON(&newUser); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}

	if newUser.Login == "" || newUser.Password == "" || newUser.GUID == "" {
		c.IndentedJSON(http.StatusUnauthorized, "")
		return
	}

	coll := client.Database("test").Collection("users")

	filter := bson.D{{"login", newUser.Login}, {"password", newUser.Password}, {"guid", newUser.GUID}}

	var result data
	err = coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.IndentedJSON(http.StatusUnauthorized, "")
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, "")
			return
		}
	}

	payload := jwt.MapClaims{
		"sub": newUser.Login,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)
	t, err := token.SignedString(jwtSecretKey)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}

	newTokens.AccessToken = t
	refreshToken, refreshErr := NewRefreshToken()

	if refreshErr != nil {
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}

	newTokens.RefreshToken = refreshToken
	hashToken, err := HashPassword(refreshToken)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}

	update := bson.D{{"$set", bson.D{{"refreshToken", hashToken}}}}
	res, err := coll.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "")
		return
	}
	fmt.Printf("Documents matched: %v\n", res)

	c.IndentedJSON(http.StatusCreated, newTokens)
}

func NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}