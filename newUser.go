package main

import (
	"context"
	"net/http"

	"github.com/beevik/guid"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateNewUser(c *gin.Context) {
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

	if newUser.Login == "" {
		c.IndentedJSON(http.StatusNotAcceptable, "")
		return
	}

	if newUser.Password == "" {
		c.IndentedJSON(http.StatusNotAcceptable, "")
		return
	}
	coll := client.Database("test").Collection("users")
	filter := bson.D{{"login", newUser.Login}}

	var result user
	err = coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			g := guid.New()
			newUser.GUID = g.String()
			coll.InsertOne(context.TODO(), newUser)
			c.IndentedJSON(http.StatusCreated, "")
			return
		} else {
			c.IndentedJSON(http.StatusInternalServerError, "")
			return
		}
	}

	c.IndentedJSON(http.StatusNotAcceptable, "")
}