package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"sister/internal/dto/requests"
	"sister/internal/models"
	"sister/pkg/database"
	"sister/pkg/mq"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	if appEnv == "development" {
		if err := godotenv.Load(".worker.env"); err != nil {
			panic("gagal memuat file .env: " + err.Error())
		}
	}

	ch := mq.Init()
	db := database.Init()
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	// notification exchange
	if err := ch.ExchangeDeclare(
		"notification-message",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		panic(err)
	}

	router.POST(
		"/global-messages", func(c *gin.Context) {
			var req requests.CreateGlobalMessage
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			data := models.Message{
				Name:    req.Name,
				Message: req.Message,
			}
			if err := db.Create(&data).Error; err != nil {
				log.Println(err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			dataStr, err := json.Marshal(data)
			if err != nil {
				log.Println(err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			// kirim notifikasi
			if err := ch.Publish(
				"notification-message",
				"",
				false,
				false,
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        dataStr,
				},
			); err != nil {
				log.Println(err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "success"})
		},
	)
	router.POST("/test-worker", func(c *gin.Context) {
		// contoh sederhana: terima body JSON
		var msg map[string]interface{}
		if err := c.BindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}

		// balikin response sederhana
		c.String(http.StatusOK, "hello from worker")
	})

	router.Run(":8082")
}
