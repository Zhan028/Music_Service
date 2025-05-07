package main

import (
	"context"
	grpcHandler "github.com/facelessEmptiness/user_service/userService/internal/delivery/grpc"
	"github.com/facelessEmptiness/user_service/userService/internal/repository"
	"github.com/facelessEmptiness/user_service/userService/internal/usecase"
	pb "github.com/facelessEmptiness/user_service/userService/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Загрузка конфигурации из переменных окружения
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	grpcPort := os.Getenv("GRPC_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenExpStr := os.Getenv("TOKEN_EXP")

	// Парсинг длительности токена
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
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Printf("Не удалось отключиться от MongoDB: %v", err)
		}
	}()

	// Пинг базы данных для проверки подключения
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Не удалось отправить пинг к MongoDB: %v", err)
	}
	log.Println("Успешно подключено к MongoDB")

	// Инициализация базы данных
	db := client.Database(dbName)

	// Инициализация репозитория
	repo := repository.NewMongoUserRepository(db)

	// Инициализация use case
	uc := usecase.NewUserUseCase(repo)

	// Инициализация gRPC обработчика (добавлен параметр JWT, если ваш обработчик поддерживает это)
	// Если ваш обработчик не принимает эти параметры, измените эту строку соответственно
	handler := grpcHandler.NewUserServiceHandler(uc, jwtSecret, tokenExp)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, handler)

	// Включение reflection для инструментов типа grpcurl
	reflection.Register(grpcServer)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Не удалось запустить прослушивание: %v", err)
	}
	log.Printf("Запуск gRPC сервера на порту %s", grpcPort)

	// Обработка корректного завершения
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		log.Println("Выключение gRPC сервера...")
		grpcServer.GracefulStop()
	}()

	// Начало обслуживания
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Не удалось обслужить: %v", err)
	}
}
