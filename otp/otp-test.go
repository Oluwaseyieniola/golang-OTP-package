package otp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

type MockKafkaWriter struct {
	Messages []kafka.Message
}

func (m *MockKafkaWriter) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	m.Messages = append(m.Messages, msgs...)
	return nil // Simulate a successful write
}

func (m *MockKafkaWriter) Close() error {
	return nil // Simulate a successful close
}

func TestOTPGenerationAndValidation(t *testing.T) {
	mockWriter := &MockKafkaWriter{}
	manager := &OTPManager{
		store:       make(map[string]OTP),
		Validity:    2 * time.Minute,
		KafkaWriter: mockWriter,
	}

	// Test numeric OTP generation
	otp, err := manager.GenerateOTP("user1@example.com", "numeric", 6)
	if err != nil {
		t.Fatalf("Failed to generate numeric OTP: %v", err)
	}
	if len(otp.Token) != 6 {
		t.Fatalf("Expected OTP of length 6, got: %s", otp.Token)
	}

	// Test alphanumeric OTP generation
	otpAlpha, err := manager.GenerateOTP("user2@example.com", "alphanumeric", 8)
	if err != nil {
		t.Fatalf("Failed to generate alphanumeric OTP: %v", err)
	}
	if len(otpAlpha.Token) != 8 {
		t.Fatalf("Expected alphanumeric OTP of length 8, got: %s", otpAlpha.Token)
	}

	// Validate valid OTP
	valid, message := manager.validate("user1@example.com", otp.Token)
	if !valid {
		t.Fatalf("Expected valid OTP, got: %s", message)
	}

	// Test expired OTP scenario
	manager.store["user1@example.com"] = OTP{
		Identifier: otp.Identifier,
		Token:      otp.Token,
		Type:       otp.Type,
		ExpiresAt:  time.Now().Add(-1 * time.Minute), // Simulate expiration
	}
	valid, message = manager.validate("user1@example.com", otp.Token)
	if valid || message != "OTP has expired!" {
		t.Fatalf("Expected expired OTP, got: %s", message)
	}

	// Check Kafka messages
	if len(mockWriter.Messages) != 2 {
		t.Fatalf("Expected 2 Kafka messages, got: %d", len(mockWriter.Messages))
	}

	// Validate Kafka message structure
	for _, msg := range mockWriter.Messages {
		if len(msg.Key) == 0 || len(msg.Value) == 0 {
			t.Fatalf("Expected non-empty Kafka message key and value, got empty key or value")
		}
	}

	// Test concurrent OTP generation
	runConcurrentOTPTests(manager, t)

	// Close the Kafka writer
	manager.Close()
}

func runConcurrentOTPTests(manager *OTPManager, t *testing.T) {
	numWorkers := 100
	errors := make(chan error, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(i int) {
			identifier := fmt.Sprintf("user%d@example.com", i)
			_, err := manager.GenerateOTP(identifier, "numeric", 6)
			errors <- err
		}(i)
	}

	for i := 0; i < numWorkers; i++ {
		err := <-errors
		if err != nil {
			t.Errorf("Failed to generate OTP in concurrent test: %v", err)
		}
	}
}

func TestInvalidOTPGeneration(t *testing.T) {
	manager := &OTPManager{
		store:       make(map[string]OTP),
		Validity:    2 * time.Minute,
		KafkaWriter: &MockKafkaWriter{},
	}

	// Invalid OTP type
	_, err := manager.GenerateOTP("invalid@example.com", "invalid_type", 6)
	if err == nil {
		t.Fatal("Expected error for invalid OTP type")
	}

	// Invalid length (negative)
	_, err = manager.GenerateOTP("invalid_length@example.com", "numeric", -1)
	if err == nil {
		t.Fatal("Expected error for negative OTP length")
	}
}

func TestCleanExpiredOTP(t *testing.T) {
	manager := &OTPManager{
		store:       make(map[string]OTP),
		Validity:    1 * time.Minute,
		KafkaWriter: &MockKafkaWriter{},
	}

	// Generate expired and valid OTPs
	manager.GenerateOTP("user1@example.com", "numeric", 6)
	manager.store["user2@example.com"] = OTP{
		Identifier: "user2@example.com",
		Token:      "expired_token",
		ExpiresAt:  time.Now().Add(-1 * time.Minute), // Simulate expired OTP
	}

	manager.CleanExpiredOTP("")

	// Assert expired OTP is removed
	if _, exists := manager.store["user2@example.com"]; exists {
		t.Fatal("Expired OTP was not cleaned up")
	}

	// Assert valid OTP remains
	if _, exists := manager.store["user1@example.com"]; !exists {
		t.Fatal("Valid OTP was incorrectly cleaned up")
	}
}
