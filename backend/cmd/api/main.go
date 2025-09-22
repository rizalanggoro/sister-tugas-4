package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"sister/internal/dto/requests"
	"sister/internal/dto/responses"
	"sister/internal/models"
	"sister/pkg/database"
	pb "sister/pkg/grpc"
	"sister/pkg/mq"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// Berarti ini code Rest API pake gin
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

	// grpc connection
	grpcConn, err := grpc.NewClient(
		os.Getenv("WORKER_GRPC_BASE_URL"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer grpcConn.Close()
	grpcMessageClient := pb.NewMessageClient(grpcConn)

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

	q, err := ch.QueueDeclare(
		"hello-queue",
		false,
		false,
		false,
		false,
		nil,
	)

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
		group.POST(
			"/grpc", func(c *gin.Context) {
				var req requests.CreateGlobalMessage
				if err := c.ShouldBindJSON(&req); err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}

				if _, err := grpcMessageClient.SendMessage(
					c, &pb.CreateMessageRequest{
						Name:    req.Name,
						Message: req.Message,
					},
				); err != nil {
					log.Println(err.Error())
					c.AbortWithStatus(http.StatusInternalServerError)
				} else {
					c.JSON(http.StatusOK, gin.H{"status": "success"})
				}
			},
		)
	}

	groupTest := router.Group("/test-worker")
	{
		// endpoint untuk mereturn string
		// sifatnya asynchronous
		groupTest.POST(
			"/mq", func(c *gin.Context) {
				msg := Message{Name: "mq", Message: "Hello from MQ!"}
				// ubah pesan ke dalam bentuk string json
				body, _ := json.Marshal(msg)
				if err := ch.Publish(
					"",
					q.Name,
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

				// tetap return string sederhana
				c.String(http.StatusAccepted, "hello from mq")
			},
		)

		// endpoint untuk mereturn string
		// sifatnya synchronous
		groupTest.POST(
			"/rest", func(c *gin.Context) {
				msg := Message{Name: "rest", Message: "Hello from Rest!"}
				// ubah pesan ke dalam bentuk string json
				body, _ := json.Marshal(msg)

				// teruskan pesan ke worker melalui rest api
				baseUrl := os.Getenv("WORKER_REST_API_BASE_URL")
				res, err := http.Post(
					fmt.Sprintf("%s/test-worker", baseUrl),
					"application/json",
					bytes.NewBuffer(body),
				)
				defer res.Body.Close()
				if err != nil {
					log.Println(err.Error())
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				// tetap return string sederhana
				c.String(http.StatusOK, "hello from rest")
			},
		)

		// endpoint untuk mereturn string
		// sifatnya synchronous
		groupTest.POST(
			"/grpc", func(c *gin.Context) {

				if _, err := grpcMessageClient.SendMessage(
					c, &pb.CreateMessageRequest{
						Name:    "grpc",
						Message: "hello from grpc",
					},
				); err != nil {
					log.Println(err.Error())
					c.AbortWithStatus(http.StatusInternalServerError)
				} else {
					c.String(http.StatusOK, "hello from grpc")
				}
			},
		)
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
