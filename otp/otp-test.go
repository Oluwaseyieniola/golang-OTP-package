package otp

import (
	"context"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	
)

func TestGenerateAndSendToKafka(t *testing.T) {
	// Set up Kafka connection details
	kafkaBroker := "localhost:9092" // Adjust to your Kafka broker if needed
	topic := "otp-test-topic"

	// Create a Kafka reader to verify the messages
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   topic,
		GroupID: "otp-test-group",
	})
	defer kafkaReader.Close()

	// Initialize OTP Manager
	manager := NewOtpManager(5, kafkaBroker, topic)

	// Generate OTP
	otpResult, err := manager.GenerateOTP("test@example.com", "numeric", 6)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// Check if OTP was generated
	if otpResult.Token == "" {
		t.Fatalf("OTP token is empty")
	}

	// Wait to allow Kafka to process the message
	time.Sleep(2 * time.Second)

	// Read the message from Kafka
	msg, err := kafkaReader.ReadMessage(context.Background())
	if err != nil {
		t.Fatalf("Failed to read message from Kafka: %v", err)
	}

	// Validate the Kafka message contents
	if string(msg.Key) != "test@example.com" {
		t.Fatalf("Expected identifier 'test@example.com', got '%s'", msg.Key)
	}

	expectedSubstring := otpResult.Token
	if !contains(string(msg.Value), expectedSubstring) {
		t.Fatalf("OTP token '%s' not found in Kafka message '%s'", expectedSubstring, msg.Value)
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}
