package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"orderkeeper/internal/cache"
	"orderkeeper/internal/handler"
	"orderkeeper/internal/kafka"
	"orderkeeper/internal/models"
	"orderkeeper/internal/repository"
	"orderkeeper/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	database, err := initDBWithRetry(5, 2*time.Second)
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	orderCache := cache.NewOrderCache()

	orderRepository := repository.NewOrderRepository(database)
	orderService := service.NewOrderService(orderRepository, orderCache)
	orderHandler := handler.NewOrderHandler(orderService)

	if err := orderService.RestoreCache(); err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}
	log.Printf("Cache restored with %d orders", orderCache.Count())

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "orders"
	}
	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "order-service-group"
	}

	kafkaConsumer := kafka.NewConsumer(
		[]string{kafkaBrokers},
		kafkaTopic,
		kafkaGroupID,
		orderService, // передаем только сервис
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("Starting Kafka consumer (brokers: %s, topic: %s, group: %s)",
			kafkaBrokers, kafkaTopic, kafkaGroupID)

		kafkaConsumer.Run(ctx)
	}()

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	r.Post("/order", orderHandler.CreateOrderHandler)
	r.Get("/order/{id}", orderHandler.GetOrderByIDHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	r.Handle("/web/*", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server started and listening on port %s", port)
		log.Printf("Web interface: http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %s, shutting down...", sig)
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	} else {
		log.Println("HTTP server exited properly")
	}

	cancel()

	select {
	case <-time.After(5 * time.Second):
		log.Println("Shutdown completed")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded")
	}
}

func initDBWithRetry(maxAttempts int, initialDelay time.Duration) (*gorm.DB, error) {
	dsn := os.Getenv("DSN")
	var dbInstance *gorm.DB
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			if err = dbInstance.AutoMigrate(
				&models.Delivery{},
				&models.Payment{},
				&models.Item{},
				&models.Order{},
			); err != nil {
				return nil, err
			}
			return dbInstance, nil
		}

		log.Printf("Attempt %d/%d: failed to connect to database: %v", attempt, maxAttempts, err)
		if attempt < maxAttempts {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * initialDelay
			log.Printf("Retrying in %v...", delay)
			time.Sleep(delay)
		}
	}

	return nil, err
}
