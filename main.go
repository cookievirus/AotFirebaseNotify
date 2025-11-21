package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type RequestBody struct {
	Availability float64 `json:"availability"`
	CartCount    int     `json:"cartcount"`
	StatusColor  string  `json:"statusColor"`
}

var fcmClient *messaging.Client

// ------------------------------
// Decode & Write Firebase Cred
// ------------------------------
func writeFirebaseCredentialFile() (string, error) {
	base64Cred := os.Getenv("FIREBASE_CRED_BASE64")
	credPath := os.Getenv("FIREBASE_CRED_FILE")
	if credPath == "" {
		credPath = "/tmp/service-account.json"
	}

	if base64Cred == "" {
		return "", fmt.Errorf("FIREBASE_CRED_BASE64 is empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(base64Cred)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %v", err)
	}

	err = os.WriteFile(credPath, decoded, 0600)
	if err != nil {
		return "", fmt.Errorf("write cred file error: %v", err)
	}

	return credPath, nil
}

func main() {
	_ = godotenv.Load()

	// ⭐ Force log to stdout (DigitalOcean reads only stdout)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("[INIT] Starting server...")

	// ------------------------------
	// STEP 1: Write Firebase credentials (AFTER container is ready)
	// ------------------------------
	credPath := "/tmp/service-account.json"
	os.Setenv("FIREBASE_CRED_FILE", credPath)

	// Async write (DO App Platform requires container be ready before writing)
	go func() {
		time.Sleep(1 * time.Second)
		_, err := writeFirebaseCredentialFile()
		if err != nil {
			log.Fatalf("[CRED ERROR] %v\n", err)
		}
		log.Println("[CRED] Firebase credential file written successfully")
	}()

	// Wait for credential file
	time.Sleep(2 * time.Second)

	// ------------------------------
	// STEP 2: Init Firebase
	// ------------------------------
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("[FIREBASE ERROR] %v\n", err)
	}

	fcmClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("[FCM ERROR] %v\n", err)
	}

	log.Println("[OK] Firebase initialized")

	// ------------------------------
	// STEP 3: Setup HTTP server
	// ------------------------------
	http.HandleFunc("/send", handleNotification)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[READY] Server POC Version 1.3 on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleNotification(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("SERVER_API_KEY")
	requestKey := r.Header.Get("X-API-Key")

	if apiKey != "" && requestKey != apiKey {
		log.Printf("[REJECTED] Invalid API Key: %s\n", requestKey)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[INCOMING] Availability: %.2f | Cart: %d | Color: %s\n",
		req.Availability, req.CartCount, req.StatusColor)

	dataPayload := map[string]string{
		"availability": fmt.Sprintf("%.2f%%", req.Availability),
		"cartcount":    fmt.Sprintf("%d", req.CartCount),
		"statusColor":  req.StatusColor,
	}

	topic := os.Getenv("FCM_TOPIC")
	if topic == "" {
		topic = "all_users"
	}

	msg := &messaging.Message{
		Data:  dataPayload,
		Topic: topic,
	}

	resp, err := fcmClient.Send(context.Background(), msg)
	if err != nil {
		log.Println("[ERROR] Sending FCM:", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("[SUCCESS] FCM Message ID: %s\n", resp)
	fmt.Fprintf(w, "Sent: %s", resp)
}
