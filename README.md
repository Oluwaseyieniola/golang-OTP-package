# OTP PACKAGE WITH KAFKA INTEGRATION
**Overview**

The OTP (One-Time Password) Package provides secure generation, validation, and transmission of OTPs via Kafka. This package is designed for systems that require robust OTP handling, with real-time delivery to Kafka message brokers for further processing or storage.

## Features
- Generate numeric or alphanumeric OTPs.
- Specify OTP validity duration in minutes.
- Store OTPs in memory with automatic expiration cleanup.
- Send OTPs to Kafka for distributed systems.
- Thread-safe OTP handling using mutex locks.
- Easy Kafka integration for event-driven architectures.


## Installation
To install this package
`go get github.com/Oluwaseyieniola/go-otp`

## Usage
**Import the package**
`import "github.com/Oluwaseyieniola/go-otp"`
**Initialize the OTP Manager**
`manager := otp.NewOtpManager(validityMinutes, kafkaBroker, topic)`

### Parameters:
- validityMinutes: Integer representing OTP validity in minutes.
- kafkaBroker: Address of the Kafka broker (e.g., "localhost:9092").
- topic: Kafka topic where OTPs will be sent (e.g., "otp-topic").

**Generate OTP**
```
otpResult, err := manager.GenerateOTP("user@example.com", "numeric", 6)
if err != nil {
    log.Fatalf("Failed to generate OTP: %v", err)
}

fmt.Printf("Generated OTP: %s\n", otpResult.Token)
```

### Parameters:
identifier: Unique user identifier (e.g., email or phone number).
otpType: Type of OTP, either "numeric" or "alphanumeric".
length: Length of the OTP (e.g., 6 for a six-digit OTP).

***Validate OTP***
```
isValid, message := manager.Validate("user@example.com", otpResult.Token)
if isValid {
    fmt.Println("OTP is valid!")
} else {
    fmt.Printf("OTP validation failed: %s\n", message)
}
```

**Clean Expired OTP**
`manager.CleanExpiredOTP()`

**Close Kafka**
`manager.Close()`

## Example
```
package main

import (
    "fmt"
    "log"
    "github.com/Oluwaseyieniola/go-otp"
)

func main() {
    kafkaBroker := "localhost:9092"
    topic := "otp-topic"
    manager := otp.NewOtpManager(5, kafkaBroker, topic) // OTP valid for 5 minutes

    otp, err := manager.GenerateOTP("user@example.com", "numeric", 6)
    if err != nil {
        log.Fatalf("Error generating OTP: %v", err)
    }
    fmt.Printf("Generated OTP: %s\n", otp.Token)

    isValid, message := manager.Validate("user@example.com", otp.Token)
    if isValid {
        fmt.Println("OTP is valid!")
    } else {
        fmt.Printf("OTP validation failed: %s\n", message)
    }

    manager.Close()
}
```
## Run Tests
`go test ./... -v`
***Ensure you have a Kafka broker running at localhost:9092 or adjust the Kafka broker address in your test configuration.***




## Error Handling
- OTP Generation Failure: Returns an error if the token cannot be generated.
- Kafka Send Failure: Logs an error if the OTP message cannot be sent to Kafka.
- Validation Errors: Returns messages like "OTP does not exist" or "OTP has expired".

## Security Considerations
Mutex Locking: Ensures thread-safe operations.
Expiration Management: OTPs automatically expire to prevent misuse.
Kafka Transmission: OTPs are securely transmitted to Kafka for reliable message handling.

## Dependencies
[Segment Kafka Go](https://github.com/segmentio/kafka-go) for Kafka integration.

## Contribution
Feel free to contribute by creating issues or submitting pull requests to improve the package.

## Contact
For issues or feature requests, please contact oluwaseyiogunjinmi@gmail.com.


