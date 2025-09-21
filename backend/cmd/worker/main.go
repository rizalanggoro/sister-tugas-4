package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"sister/internal/dto/requests"
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
		if err := godotenv.Load(".worker.env"); err != nil {
			panic("gagal memuat file .env: " + err.Error())
		}
	}

	ch := mq.Init()
	db := database.Init()

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

	if err := ch.Qos(
		1,
		0,
		false,
	); err != nil {
		panic(err)
	}

	messages, err := ch.Consume(
		globalMessageQueue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

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

	var forever chan struct{}
	go func() {
		for message := range messages {
			log.Printf("Received a message: %s", message.Body)

			// simpan message ke database untuk riwayat
			var body requests.CreateGlobalMessage
			if err := json.Unmarshal(message.Body, &body); err != nil {
				log.Println(err)
				message.Nack(false, false)
				continue
			}

			data := models.Message{
				Name:    body.Name,
				Message: body.Message,
			}
			if err := db.Create(&data).Error; err != nil {
				log.Println(err.Error())
				message.Nack(false, true)
				continue
			}

			dataStr, err := json.Marshal(data)
			if err != nil {
				log.Println(err.Error())
				message.Nack(false, true)
				continue
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
				message.Nack(false, true)
				continue
			}

			// ack message
			message.Ack(false)

			log.Println("success")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
