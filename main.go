package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

func main() {
	log.Println("CUPS Monitor starting...")

	// Get config from environment
	cupsURL := os.Getenv("CUPS_URL")
	if cupsURL == "" {
		cupsURL = "http://localhost:631"
	}

	pagerDutyKey := os.Getenv("PAGERDUTY_ROUTING_KEY")
	if pagerDutyKey == "" {
		log.Fatal("PAGERDUTY_ROUTING_KEY required")
	}

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	// Check every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Printf("Monitoring %s", cupsURL)

	for range ticker.C {
		if severity, msg := checkCUPS(cupsURL); severity != "" {
			sendAlert(pagerDutyKey, msg, severity)
		}
	}
}

// checkCUPS checks if CUPS is responding
// Returns severity ("critical" or "error") and message if unhealthy, empty if healthy
func checkCUPS(url string) (string, string) {
	resp, err := http.Get(url)
	if err != nil {
		return "critical", "CUPS service is down"
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return "error", "CUPS queue not accepting jobs"
	}
	return "", ""
}

// sendAlert sends alert to PagerDuty
func sendAlert(key, message, severity string) {
	payload := map[string]interface{}{
		"routing_key":  key,
		"event_action": "trigger",
		"dedup_key":    "cups-monitor",
		"payload": map[string]interface{}{
			"summary":  message,
			"severity": severity,
		},
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(pagerDutyEventsURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Alert failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 202 {
		log.Println("Alert sent")
	} else {
		log.Printf("Alert failed: status %d", resp.StatusCode)
	}
}
