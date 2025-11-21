# Go FCM Notification Service & Android Client POC Version 1.3

A Proof of Concept (POC) system for sending Push Notifications via a Golang REST API wrapper to an Android (Kotlin) application using Firebase Cloud Messaging (FCM).

This project is designed to abstract the complexity of Firebase Authentication from the client, allowing notifications to be sent via a simple cURL command secured by an API Key.

## 🚀 Features

* **Golang Backend:**
  * REST API endpoint to receive notification requests.
  * Secured with a custom `X-API-Key`.
  * Environment variable configuration (`.env`).
  * Integration with Firebase Admin SDK.
* **Android Client:**
  * Built with Kotlin.
  * Handles Data Payloads in background/foreground.
  * Automatically subscribes to a global topic (`all_users`).
  * Displays system notifications with custom data.

## 📂 Project Structure

```
.
├── main.go                  # Golang Server Entry point
├── go.mod                   # Go module definitions
├── .env                     # Environment variables (Not committed)
├── service-account.json     # Firebase Admin Key (Not committed)
├── .gitignore               # Git ignore rules
├── README.md                # Project documentation
└── app/                     # Android Application Source
    ├── src/main/java/...    # Kotlin Source Code
    └── google-services.json # Android Firebase Config (Not committed)
```

## 🛠 Prerequisites

- Go (1.18 or higher)
- Android Studio (Latest version)
- Firebase Project:
  - Generate `service-account.json` (Project Settings > Service Accounts).
  - Generate `google-services.json` (Project Settings > General > Add Android App).

## ⚙️ Setup & Installation

### 1. Backend (Golang)

#### Install dependencies:

```bash
go mod tidy
```

#### Configure Environment (.env):

```
PORT=8080
SERVER_API_KEY=DaM3ca1asw5sosplTaBusosT9fiT9yIk
FIREBASE_CRED_FILE=service-account.json
FCM_TOPIC=all_users
```

#### Run Server:

```bash
go run main.go
```

---

### 2. Client (Android)

- Open project in Android Studio.
- Place `google-services.json` into `app/`.
- Sync Gradle.
- Run on emulator or physical device.
- Logcat should show: `Subscribed to 'all_users'`.

---

## 📡 Usage (Sending Notifications)

Send notification using cURL.

### Endpoint
`POST /send`

### Headers
`X-API-Key` must match `.env`.

### Example:

```bash
curl -X POST http://localhost:8080/send  -H "Content-Type: application/json"  -H "X-API-Key: DaM3ca1asw5sosplTaBusosT9fiT9yIk"  -d '{"availability": 99.99, "cartcount": 5, "statusColor": "#00FF00"}'
```

### Request Body Parameters

| Parameter     | Type   | Description                                   |
|--------------|--------|-----------------------------------------------|
| availability | float  | Stock availability %                           |
| cartcount    | int    | Number of items                                |
| statusColor  | string | Hex color code                                 |

## 📝 Notes

- Android app processes Data Messages regardless of foreground/background.
- Server sends to topic `all_users`.
