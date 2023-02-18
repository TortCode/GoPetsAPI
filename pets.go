package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
    _ "time"
)

var coll *mongo.Collection

type Pet struct {
    ID     primitive.ObjectID `json:"id" bson:"_id"`
    Name   string             `json:"name" bson:"name"`
    Owner  string             `json:"owner" bson:"owner"`
    DOB    primitive.DateTime `json:"birthdate" bson:"birthdate"`
    Type   string             `json:"type" bson:"type"`
    Height int                `json:"height" bson:"height"`
    Width  int                `json:"width" bson:"width"`
    Toy    string             `json:"favtoy" bson:"favtoy"`
}

func create_pet(c *gin.Context) {
    var pet Pet
    if err := c.BindJSON(&pet); err != nil {
        return
    }

    // generate object id
    pet.ID = primitive.NewObjectID()

    // insert element
    _, insert_err := coll.InsertOne(context.TODO(), pet)
    if insert_err != nil {
        return
    }

    c.JSON(http.StatusCreated, pet)
}

func delete_pet(c *gin.Context) {
    id, objectid_err := primitive.ObjectIDFromHex(c.Param("id"))
    if objectid_err != nil {
        panic(objectid_err)
    }
    filter := bson.D{{"_id", id}}
    result, delete_err := coll.DeleteOne(context.TODO(), filter)
    if delete_err != nil {
        panic(delete_err)
    }
    if result.DeletedCount == 0 {
        c.Status(http.StatusNotFound)
    } else {
        c.Status(http.StatusNoContent)
    }
}

func update_pet(c *gin.Context) {
    id, objectid_err := primitive.ObjectIDFromHex(c.Param("id"))
    if objectid_err != nil {
        panic(objectid_err)
    }

    var updates bson.M
    if err := c.BindJSON(&updates); err != nil {
        panic(err)
    }
   
    update_request := bson.D{{"$set", updates}}
    result, err := coll.UpdateByID(context.TODO(), id, update_request)
    if err != nil {
        panic(err)
    }

    if result.MatchedCount == 0 {
        c.Status(http.StatusNotFound)
    } else {
        c.Status(http.StatusNoContent)
    }

}

func retrieve_pets(c *gin.Context) {
    var result []Pet 
    cursor, err := coll.Find(context.TODO(), bson.D{})
    if err != nil { 
        panic(err) 
    }

    if err := cursor.All(context.TODO(), &result); err != nil {
        panic(err)
    }

    c.JSON(http.StatusOK, result)
}

func main() {
    godotenv.Load()
    
    mongo_username := os.Getenv("MONGODB_USERNAME")
    mongo_password := url.QueryEscape(os.Getenv("MONGODB_PASSWORD"))
    mongo_cluster := os.Getenv("MONGODB_CLUSTER")
    mongo_uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/phonebookApp?retryWrites=true&w=majority",
        mongo_username, mongo_password, mongo_cluster)

    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongo_uri))

    if err != nil {
        panic(err)
    }

    //disconnect on completion
    defer func() {
        if err := client.Disconnect(context.TODO()); err != nil {
            panic(err)
        }
    }()

    coll = client.Database("pets_profiles").Collection("pets")

    router := gin.Default()

    pets_api := router.Group("/api/pets")
    {
        pets_api.GET("", retrieve_pets)
        pets_api.POST("", create_pet)
        pets_api.DELETE("/:id", delete_pet)
        pets_api.PATCH("/:id", update_pet)
    }

    router.Run("localhost:3000")
}
