package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"sister/pkg/mq"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleWebSocket(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	log.Println("clients", clients)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			break
		}
	}
}

func main() {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	if appEnv == "development" {
		if err := godotenv.Load(".notification.env"); err != nil {
			panic("gagal memuat file .env: " + err.Error())
		}
	}

	ch := mq.Init()
	router := gin.Default()

	router.GET("/ws", handleWebSocket)

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

	queue, err := ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	if err := ch.QueueBind(
		queue.Name,
		"",
		"notification-message",
		false,
		nil,
	); err != nil {
		panic(err)
	}

	messages, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for message := range messages {
			log.Printf(" [x] %s", message.Body)

			clientsMu.Lock()
			for client := range clients {
				if err := client.WriteMessage(websocket.TextMessage, message.Body); err != nil {
					client.Close()
					delete(clients, client)
				}
			}
			clientsMu.Unlock()
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	router.Run(":8081")
}
