package mq

import (
	"fmt"
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	channelInstance *amqp.Channel
	once            sync.Once
)

func Init() *amqp.Channel {
	once.Do(
		func() {
			mqHost := os.Getenv("MQ_HOST")
			mqPort := os.Getenv("MQ_PORT")
			mqUser := os.Getenv("MQ_USER")
			mqPassword := os.Getenv("MQ_PASSWORD")

			url := fmt.Sprintf(
				"amqp://%s:%s@%s:%s/",
				mqUser, mqPassword, mqHost, mqPort,
			)

			log.Println(url)

			connection, err := amqp.Dial(url)
			if err != nil {
				panic(err)
			}

			channel, err := connection.Channel()
			if err != nil {
				panic(err)
			}

			channelInstance = channel
		},
	)

	return channelInstance
}
