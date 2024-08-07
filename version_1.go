package main

import (
    "encoding/json"
    "io"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
)

type RequestBody struct {
    Region string `json:"region"`
}

func main() {
    r := gin.Default()

    r.POST("/get-data", func(c *gin.Context) {
        var requestBody RequestBody

        if err := c.ShouldBindJSON(&requestBody); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
            return
        }

        var filename string
        switch requestBody.Region {
        case "ap":
            filename = "data/ap-phase-1.json"
        case "ka":
            filename = "data/ka-phase-1.json"
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid region"})
            return
        }

        file, err := os.Open(filename)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read file"})
            return
        }
        defer file.Close()

        jsonData, err := io.ReadAll(file)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading file"})
            return
        }

        var data interface{}
        if err := json.Unmarshal(jsonData, &data); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unmarshaling JSON"})
            return
        }

        c.JSON(http.StatusOK, data)
    })

    r.Run(":8085")
}
