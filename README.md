# 🎵 Music_Service

A scalable microservices-based **Music Streaming Platform**, developed in Go. The system supports user registration, track management, playlist creation, and unified access through an API Gateway. The services communicate via gRPC and integrate Kafka for messaging and Redis for caching. All data is stored in **MongoDB**, with full support for migrations, transactions, and testing.

---

## 📁 Microservices Overview

### 1. 🧍 userService
- Handles user registration and login
- Password hashing with bcrypt
- JWT token generation for authentication
- Email notifications via SMTP
- Stores user data in MongoDB

### 2. 🎵 track-service
- Manages songs (add, update, delete)
- Stores metadata: artist, genre, duration, etc.
- Supports audio file upload and access
- Uses MongoDB for track storage

### 3. 📂 playlistService
- Allows users to create, edit, and delete playlists
- Add/remove tracks to playlists
- Stores playlists in MongoDB

### 4. 🌐 api_gateway
- Routes and connects all microservices
- Handles JWT authentication
- Unified gRPC API entry point

---

## 🧰 Technologies Used

- **Language:** Go (Golang)
- **API Protocol:** gRPC
- **Message Broker:** Kafka
- **Cache:** Redis
- **Database:** MongoDB for all services
- **Migrations:** Custom scripts / tools for MongoDB
- **Authentication:** JWT
- **Email:** Gomail via SMTP

---

## 🧪 Testing

- ✅ Unit Testing for service logic
- 🔄 Integration Testing for full service interactions
- Covers database operations, gRPC communication, and message processing

---

## 🗄 Database

- **MongoDB** used for storing user, track, and playlist data
- Transactions are used where applicable (multi-document)
- Migrations managed through custom setup or third-party tools (e.g., Mongock)

---

## 🚀 Running the Project

> Each microservice runs independently. Ensure MongoDB, Kafka, and Redis are running.

```bash
# Run userService
go run ./cmd/main.go
