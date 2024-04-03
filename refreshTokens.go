package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RefreshTokens(c *gin.Context) {
	responce := c.GetHeader("Authorization")

	if c.GetHeader("Authorization") == "" {
		c.IndentedJSON(http.StatusUnauthorized, "")
		return
	}

	header := strings.Fields(responce)

	if header[0] != "Bearer" {
		c.IndentedJSON(http.StatusUnauthorized, "")
		return
	}
	if header[1] == "" {
		c.IndentedJSON(http.StatusUnauthorized, "")
		return
	}

	token, err := jwt.Parse(header[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
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

		var r refreshData
		if err := c.BindJSON(&r); err != nil {
			c.IndentedJSON(http.StatusUnauthorized, "")
			return
		}

		coll := client.Database("test").Collection("users")
		filter := bson.D{{"login", claims["sub"]}}

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

		check := CheckPasswordHash(r.Refresh, result.RefreshToken)

		if check {
			var newTokens tokens
			payload := jwt.MapClaims{
				"sub": claims["sub"],
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
			return
		} else {
			c.IndentedJSON(http.StatusUnauthorized, "")
			return
		}
	} else {
		c.IndentedJSON(http.StatusUnauthorized, err)
		return
	}
}