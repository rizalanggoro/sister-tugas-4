package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"sister/internal/dto/requests"
	"sister/internal/dto/responses"
	"sister/internal/models"
	"sister/pkg/database"
	"sister/pkg/mq"
)

func main() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	if appEnv == "development" {
		if err := godotenv.Load(".api.env"); err != nil {
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

	globalMessageQueue, err := ch.QueueDeclare(
		"global-message",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	router.POST(
		"/global-messages", func(c *gin.Context) {
			var req requests.CreateGlobalMessage
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			// ubah pesan ke dalam bentuk string json
			body, err := json.Marshal(req)
			if err != nil {
				log.Println(err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			if err := ch.Publish(
				"",
				globalMessageQueue.Name,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			); err != nil {
				log.Println(err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
		},
	)
	router.GET(
		"/global-messages", func(c *gin.Context) {
			var messages []models.Message
			if err := db.Order("created_at asc").Limit(100).Find(&messages).Error; err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.JSON(
				http.StatusOK, responses.GetAllGlobalMessages{
					Messages: messages,
				},
			)
		},
	)

	router.Run(":8080")
}
