package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type RequestBody struct {
    Region string `json:"region"`
}

// MongoDB middleware to attach the client to the context
func mongoMiddleware(client *mongo.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("mongoClient", client)
        c.Next()
    }
}

func main() {
    // Load .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Establish MongoDB connection
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        log.Fatalf("MongoDB URI is not set")
    }

    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatalf("Could not connect to MongoDB: %v", err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatalf("Could not ping MongoDB: %v", err)
    }

    // Ensure client disconnects when the application closes
    defer client.Disconnect(context.TODO())

    r := gin.Default()

    // Apply MongoDB middleware
    r.Use(mongoMiddleware(client))

    r.POST("/get-data", func(c *gin.Context) {
        var requestBody RequestBody

        if err := c.ShouldBindJSON(&requestBody); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        // Retrieve the MongoDB client from context
        mongoClient, exists := c.Get("mongoClient")
        if !exists {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB client not found"})
            return
        }

        client := mongoClient.(*mongo.Client)

        var collection *mongo.Collection
        switch requestBody.Region {
        case "ap":
            collection = client.Database("andhra_pradesh").Collection("phase_1")
        case "ka":
            collection = client.Database("karnataka").Collection("phase_1")
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region"})
            return
        }

        var results []map[string]interface{}
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        cursor, err := collection.Find(ctx, map[string]interface{}{})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data from MongoDB"})
            return
        }
        defer cursor.Close(ctx)

        if err = cursor.All(ctx, &results); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding MongoDB data"})
            return
        }

        c.JSON(http.StatusOK, results)
    })

    r.Run(":8085")
}
