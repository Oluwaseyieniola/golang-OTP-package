package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	otp "github.com/Oluwaseyieniola/go-otp/otp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func mustEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func main() {
	kafka := mustEnv("KAFKA_BROKERS", "localhost:9092")
	topic := mustEnv("KAFKA_TOPIC", "otp-topic")
	vm := mustEnv("OTP_VALID_MINUTES", "5")
	mins, _ := strconv.Atoi(vm)

	mgr := otp.NewOtpManager(mins, kafka, topic)
	defer mgr.Close()

	// healthz & simple HTTP endpoints for demo
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	http.Handle("/metrics", promhttp.Handler()) // Prometheus

	// (optional) simple generate endpoint: /generate?id=email&type=numeric&len=6
	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		typ := r.URL.Query().Get("type")
		l := r.URL.Query().Get("len")
		n, _ := strconv.Atoi(l)
		token, err := mgr.GenerateOTP(id, typ, n)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(token.Token))
	})

	log.Println("otp-service up on :8080, metrics on :9090")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
