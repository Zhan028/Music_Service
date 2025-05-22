package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcHandler "github.com/Zhan028/Music_Service/userService/internal/delivery/grpc"
	"github.com/Zhan028/Music_Service/userService/internal/infrastructure/email"
	"github.com/Zhan028/Music_Service/userService/internal/repository"
	"github.com/Zhan028/Music_Service/userService/internal/usecase"
	pb "github.com/Zhan028/Music_Service/userService/proto"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Конфигурация сервиса
	mongoURI := "mongodb://localhost:27017"
	dbName := "ap2"
	grpcPort := "50053"
	jwtSecret := "your-secret-key"
	tokenExpStr := "24h"

	// Конфигурация SMTP (здесь рекомендуется использовать переменные окружения)
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	smtpUser := "zhanbatyrmolkryt@gmail.com"
	smtpPassword := "yesz ypaf uxcd gmdz"

	// Парсинг продолжительности токена
	tokenExp, err := time.ParseDuration(tokenExpStr)
	if err != nil {
		log.Fatalf("Неверная длительность токена: %v", err)
	}

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Printf("Не удалось отключиться от MongoDB: %v", err)
		}
	}()

	// Проверка подключения к MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Не удалось отправить пинг к MongoDB: %v", err)
	}
	log.Println("Успешно подключено к MongoDB")

	// Инициализация репозитория и базы данных
	db := client.Database(dbName)
	repo := repository.NewMongoUserRepository(db)

	// Инициализация email-отправщика
	emailSender := email.NewGomailSender(smtpUser, smtpHost, smtpUser, smtpPassword, smtpPort)

	// Инициализация бизнес-логики (use case)
	uc := usecase.NewUserUseCase(repo, emailSender)

	// Инициализация gRPC обработчика
	handler := grpcHandler.NewUserServiceHandler(uc, jwtSecret, tokenExp)

	// Настройка gRPC сервера
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Не удалось запустить прослушивание: %v", err)
	}
	log.Printf("Запуск gRPC сервера на порту %s", grpcPort)

	// Обработка сигнала завершения
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		log.Println("Выключение gRPC сервера...")
		grpcServer.GracefulStop()
	}()

	// Запуск сервера
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка запуска gRPC сервера: %v", err)
	}
}
