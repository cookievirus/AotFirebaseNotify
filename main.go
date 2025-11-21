package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os" // [เพิ่ม] ต้องใช้ os เพื่ออ่าน environment variable

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/joho/godotenv" // [เพิ่ม]
	"google.golang.org/api/option"
)

type RequestBody struct {
	Availability float64 `json:"availability"`
	CartCount    int     `json:"cartcount"`
	StatusColor  string  `json:"statusColor"`
}

var fcmClient *messaging.Client

func main() {
	// [เพิ่ม] 1. โหลดค่าจากไฟล์ .env (ถ้าหาไม่เจอให้แจ้งเตือน แต่ไม่ถึงกับพังถ้าเรา set environment ไว้ใน OS แล้ว)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, checking OS environment variables")
	}

	// [แก้ไข] 2. อ่านค่า Config จาก Environment Variable
	credFile := os.Getenv("FIREBASE_CRED_FILE")
	if credFile == "" {
		credFile = "service-account.json" // Default fallback
	}

	opt := option.WithCredentialsFile(credFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	fcmClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}

	http.HandleFunc("/send", handleNotification)

	// [แก้ไข] 3. อ่าน Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server POC Version 1.3") // [เพิ่ม] แสดงเวอร์ชันของเซิร์ฟเวอร์
	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleNotification(w http.ResponseWriter, r *http.Request) {
	// [แก้ไข] 4. อ่าน API Key จาก Env
	apiKey := os.Getenv("SERVER_API_KEY")

	requestKey := r.Header.Get("X-API-Key")
	// เช็คว่ามีการตั้งค่า key ไหม และ key ตรงกันไหม
	if apiKey != "" && requestKey != apiKey {
		log.Printf("[REJECTED] Invalid API Key attempt: %s\n", requestKey)
		http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized)
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

	log.Printf("[INCOMING] Availability: %.2f | Cart: %d | Color: %s\n", req.Availability, req.CartCount, req.StatusColor)

	dataPayload := map[string]string{
		"availability": fmt.Sprintf("%.2f%%", req.Availability),
		"cartcount":    fmt.Sprintf("%d", req.CartCount),
		"statusColor":  req.StatusColor,
	}

	// [แก้ไข] 5. อ่าน Topic จาก Env
	topic := os.Getenv("FCM_TOPIC")
	if topic == "" {
		topic = "all_users" // Default fallback
	}

	message := &messaging.Message{
		Data:  dataPayload,
		Topic: topic,
	}

	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Println("Error sending message:", err)
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	log.Printf("[SUCCESS] Sent to Firebase. Message ID: %s\n", response)
	fmt.Fprintf(w, "Successfully sent message: %s", response)
}
