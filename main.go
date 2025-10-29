package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// State tracks current health status
var isHealthy = true

func main() {
	log.Println("CUPS Monitor starting...")

	// Load config from .env file
	config := loadEnv(".env")

	cupsURL := config["CUPS_URL"]
	if cupsURL == "" {
		cupsURL = "http://localhost:631"
	}

	pagerDutyKey := config["PAGERDUTY_ROUTING_KEY"]
	if pagerDutyKey == "" {
		log.Fatal("PAGERDUTY_ROUTING_KEY required in .env")
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
		checkAndAlert(cupsURL, pagerDutyKey)
	}
}

// loadEnv reads .env file and returns a map of key-value pairs
func loadEnv(filename string) map[string]string {
	config := make(map[string]string)

	data, err := os.ReadFile(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to read .env: %v", err)
		}
		return config
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	return config
}

// checkAndAlert checks CUPS health and sends alerts on state changes
func checkAndAlert(url, key string) {
	healthy, severity, msg := checkCUPS(url)

	if healthy {
		// Service is healthy
		if !isHealthy {
			// State change: unhealthy → healthy
			log.Println("CUPS recovered")
			sendAlert(key, url, "CUPS recovered", "info", "resolve")
			isHealthy = true
		}
	} else {
		// Service is unhealthy
		if isHealthy {
			// State change: healthy → unhealthy
			log.Printf("CUPS down: %s", msg)
			sendAlert(key, url, msg, severity, "trigger")
			isHealthy = false
		}
	}
}

// checkCUPS checks if CUPS is responding
// Returns healthy status, severity, and message
func checkCUPS(url string) (bool, string, string) {
	resp, err := http.Get(url)
	if err != nil {
		return false, "critical", "CUPS service is down"
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return false, "error", "CUPS queue not accepting jobs"
	}
	return true, "", ""
}

// sendAlert sends alert to PagerDuty
func sendAlert(key, source, message, severity, action string) {
	payload := map[string]interface{}{
		"routing_key":  key,
		"event_action": action,
		"dedup_key":    "cups-monitor",
		"payload": map[string]interface{}{
			"summary":  message,
			"source":   source,
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

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 202 {
		log.Println("Alert sent")
	} else {
		log.Printf("Alert failed: status %d - %s", resp.StatusCode, string(body))
	}
}
