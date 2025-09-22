package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	group := router.Group("/global-messages")
	{
		// endpoint untuk membuat pesan menggunakan message queue
		// sifatnya asynchronous
		group.POST(
			"/mq", func(c *gin.Context) {
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

		// endpoint untuk membuat pesan menggunakan rest api
		// sifatnya synchronous
		group.POST(
			"/rest", func(c *gin.Context) {
				var req requests.CreateGlobalMessage
				if err := c.ShouldBindJSON(&req); err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}

				body, err := json.Marshal(req)
				if err != nil {
					log.Println(err.Error())
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				// teruskan pesan ke worker melalui rest api
				baseUrl := os.Getenv("WORKER_REST_API_BASE_URL")
				res, err := http.Post(
					fmt.Sprintf("%s/global-messages", baseUrl),
					"application/json",
					bytes.NewBuffer(body),
				)
				defer res.Body.Close()
				if err != nil {
					log.Println(err.Error())
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				c.JSON(http.StatusOK, gin.H{"status": "success"})
			},
		)

		// endpoint untuk membuat pesan menggunakan grpc
		// sifatnya synchronous
		group.POST("/grpc")
	}

	router.GET(
		"/global-messages", func(c *gin.Context) {
			var messages []models.Message
			if err := db.Order("created_at desc").Limit(100).Find(&messages).Error; err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
				messages[i], messages[j] = messages[j], messages[i]
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
