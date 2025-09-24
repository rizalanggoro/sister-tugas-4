package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	"sister/internal/models"
	"sister/pkg/database"
	pb "sister/pkg/grpc"
	"sister/pkg/mq"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	ch *amqp.Channel
)

type MessageServer struct {
	pb.UnimplementedMessageServer
	mode string
}

func (s *MessageServer) SendMessage(_ context.Context, req *pb.CreateMessageRequest) (
	*pb.CreateMessageResponse, error,
) {
	data := models.Message{
		Name:    req.Name,
		Message: req.Message,
	}
	if err := db.Create(&data).Error; err != nil {
		return nil, err
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &pb.CreateMessageResponse{
		Id: uint64(data.ID),
	}, nil
}

func (s *MessageServer) SendDummyMessage(_ context.Context, req *pb.CreateDummyMessageRequest) (
	*pb.CreateDummyMessageResponse, error,
) {
	return &pb.CreateDummyMessageResponse{
		Message: req.Message,
	}, nil
}

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

	ch = mq.Init()
	db = database.Init()

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

	lis, err := net.Listen("tcp", ":8083")
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	pb.RegisterMessageServer(server, &MessageServer{})
	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}
