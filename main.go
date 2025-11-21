package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings" // สำหรับ replacing \n ถ้าจำเป็น

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// [NEW] Struct สำหรับรับ Payload แบบง่าย: {"displayText": "HELLO WORLD"}
type RequestBody struct {
	DisplayText string `json:"displayText"`
}

var fcmClient *messaging.Client

// ----------------------------------------------------
// STEP 1: Main Function (Load Config and Init Firebase)
// ----------------------------------------------------
func main() {
	_ = godotenv.Load() // Load local .env file
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Println("[INIT] Starting server...")

	// 1. Get Base64 String from Environment
	credBase64 := os.Getenv("FIREBASE_CRED_BASE64")
	if credBase64 == "" {
		log.Fatalf("[CRED ERROR] FIREBASE_CRED_BASE64 environment variable is empty. Cannot initialize Firebase.")
	}

	// 2. Decode Base64 (Handling potential newlines if input came from a CLI pipe)
	// **Note: This replaces the file writing logic**
	credBase64 = strings.ReplaceAll(credBase64, "\n", "") // Clean up newlines for robustness
	
	decodedCreds, err := base64.StdEncoding.DecodeString(credBase64)
	if err != nil {
		log.Fatalf("[CRED ERROR] Base64 decode error: %v", err)
	}

	// 3. Init Firebase directly from JSON bytes
	opt := option.WithCredentialsJSON(decodedCreds)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("[FIREBASE ERROR] Failed to initialize app: %v\n", err)
	}

	fcmClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("[FCM ERROR] Failed to get Messaging client: %v\n", err)
	}

	log.Println("[OK] Firebase initialized successfully from secure environment variable.")

	// ------------------------------
	// STEP 2: Setup HTTP server
	// ------------------------------
	http.HandleFunc("/send", handleNotification)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[READY] Server POC Version 1.5 on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ----------------------------------------------------
// STEP 3: Handle Notification Request
// ----------------------------------------------------
func handleNotification(w http.ResponseWriter, r *http.Request) {
	// 1. API Key Check
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

	// 2. Decode JSON (Simplified Payload)
	var req RequestBody // ใช้ struct ใหม่
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 3. Log Payload
	log.Printf("[INCOMING] DisplayText: %s\n", req.DisplayText)

	// 4. Create Data Payload
	dataPayload := map[string]string{
		"displayText": req.DisplayText, // ใช้ DisplayText
	}

	topic := os.Getenv("FCM_TOPIC")
	if topic == "" {
		topic = "all_users"
	}

	msg := &messaging.Message{
		Data:  dataPayload,
		Topic: topic,
	}

	// 5. Send FCM
	resp, err := fcmClient.Send(context.Background(), msg)
	if err != nil {
		log.Println("[ERROR] Sending FCM:", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("[SUCCESS] FCM Message ID: %s\n", resp)
	fmt.Fprintf(w, "Sent: %s", resp)
}