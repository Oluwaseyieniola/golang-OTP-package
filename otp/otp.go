package otp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type OTP struct {
	Identifier string
	Token      string
	Type       string
	ExpiresAt  time.Time
}

type OTPManager struct {
	store    map[string]OTP
	mu       sync.Mutex
	Validity time.Duration
	KafkaWriter  *kafka.Writer
}

func NewOtpManager(validityMinutes int, kafkaBroker, topic string) *OTPManager {
	return &OTPManager{
		store:    make(map[string]OTP),
		Validity: time.Duration(validityMinutes) * time.Minute,
		KafkaWriter: &kafka.Writer{
			Addr: kafka.TCP(kafkaBroker),
			Topic: topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (manager *OTPManager) GenerateOTP(identifier string, otpType string, length int) (OTP, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	var token string
	var err error

	switch otpType {
	case "numeric":
		token, err = generateNumericToken(length)

	case "alphanumeric":
		token, err = generateAlphanumericToken(length)
	}

	if err != nil {
		return OTP{}, err
	}

	// otp object
	NewOtp := OTP{
		Identifier: identifier,
		Token:      token,
		Type:       otpType,
		ExpiresAt:  time.Now().Add(manager.Validity),
	}
	manager.store[identifier] = NewOtp

	go func(){
		message:= fmt.Sprintf(`{"identifier": "%s" , "token": "%s" ,"type": "%s", "expiresAt": "%s"}`, NewOtp.Identifier, NewOtp.Token, NewOtp.Type, NewOtp.ExpiresAt, NewOtp.ExpiresAt.Format(time.RFC3339) )

		err:= manager.KafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Key: []byte(NewOtp.Identifier),
			Value: []byte(message),
		})

		if err != nil{
			log.Printf("Failed to send OTP to Kafka: %v\n", err)
		}else{
			log.Printf("OTP sent successfully to Kafka\n", message)
		}
	}()

	return NewOtp, nil
}
// generate numeric tokens
func generateNumericToken(length int) (string, error) {
	var result string

	for i := 0; i <= length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", nil
		}

		result += fmt.Sprintf("%d", num)
	}
	return result, nil
}
// generate alphanumeric tokens
func generateAlphanumericToken(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes)[:length], nil
}

// validate OTP provided
func (manager *OTPManager) validate(identifier string, token string) (bool, string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	otp, exists := manager.store[identifier]
	if !exists {
		return false, "OTP does not exist"
	}
	// to check if otp has expired

	if time.Now().After(otp.ExpiresAt) {
		return false, "OTP has expired!"
	}

	delete(manager.store, identifier)
	return true, "OTP is valid."
}

//  clean up old tokens

func (manager *OTPManager) CleanExpiredOTP(identifier string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for identifier, otp := range manager.store {
		if time.Now().After(otp.ExpiresAt) {
			delete(manager.store, identifier)
		}
	}

}

func(manager *OTPManager) Close(){
	manager.KafkaWriter.Close()
	log.Println("Kafka writer closed")
}




