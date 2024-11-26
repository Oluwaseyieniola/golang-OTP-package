package otp

import (
	"testing"
)

func TestGenerateAndValidate(t *testing.T) {
	manager := NewOtpManager(10) // OTP valid for 10 minutes

	// Generate an OTP
	otp, err := manager.GenerateOTP("test@example.com", "numeric", 6)
	if err != nil || otp.Token == "" {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// Validate the OTP
	valid, message := manager.validate("test@example.com", otp.Token)
	if !valid {
		t.Fatalf("Expected valid OTP, got: %s", message)
	}

	// Test expired OTP
	manager.CleanExpiredOTP(otp.Identifier) // Manually trigger cleanup to simulate expiration
	valid, message = manager.validate("test@example.com", otp.Token)
	if valid || message != "OTP expired" {
		t.Fatalf("Expected expired OTP, got: %s", message) 
	}
}
