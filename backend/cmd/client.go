package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "sister/pkg/grpc" // sesuaikan dengan path proto kamu

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// struktur pesan
type Message struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func testMQ() {
	conn, err := amqp.Dial("amqp://rizalanggoro:130603@localhost:5672/") // sesuaikan
	if err != nil {
		log.Fatal("Gagal konek ke RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Gagal buka channel:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("hello-queue", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Gagal declare queue:", err)
	}

	msg := Message{Name: "mq", Message: "Halo dari MQ!"}
	body, _ := json.Marshal(msg)

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		log.Fatal("Gagal publish ke MQ:", err)
	}

	fmt.Println("✅ MQ message terkirim:", string(body))
}

func testREST() {
	baseURL := os.Getenv("WORKER_REST_API_BASE_URL") // contoh: http://localhost:8082
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}

	msg := Message{Name: "rest", Message: "Halo dari REST!"}
	body, _ := json.Marshal(msg)

	res, err := http.Post(fmt.Sprintf("%s/test-worker", baseURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatal("Gagal kirim REST:", err)
	}
	defer res.Body.Close()

	fmt.Println("✅ REST response status:", res.Status)
}

func testGRPC() {
	conn, err := grpc.Dial("localhost:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Gagal konek gRPC:", err)
	}
	defer conn.Close()

	client := pb.NewMessageClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SendMessage(ctx, &pb.CreateMessageRequest{
		Name:    "grpc",
		Message: "Halo dari gRPC!",
	})
	if err != nil {
		log.Fatal("Gagal kirim gRPC:", err)
	}

	fmt.Println("✅ gRPC response ID:", resp.Id)
}

func main() {
	fmt.Println("=== Test Worker Communication ===")
	testMQ()
	testREST()
	testGRPC()
}
