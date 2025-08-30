// @title OrderKeeper API
// @version 1.0
// @description API for managing orders
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "orderkeeper/docs"
	"orderkeeper/internal/cache"
	"orderkeeper/internal/db"
	"orderkeeper/internal/handler"
	"orderkeeper/internal/kafka"
	"orderkeeper/internal/repository"
	"orderkeeper/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type Config struct {
	Port         string
	DSN          string
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
}

func NewConfig() (*Config, error) {
	log.Println("Loading configuration...")
	cfg := &Config{
		Port:         os.Getenv("PORT"),
		DSN:          os.Getenv("DSN"),
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
		KafkaTopic:   os.Getenv("KAFKA_TOPIC"),
		KafkaGroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DSN == "" {
		return nil, errors.New("DSN environment variable is not set")
	}
	if cfg.KafkaBrokers == "" {
		return nil, errors.New("KAFKA_BROKERS environment variable is not set")
	}
	if cfg.KafkaTopic == "" {
		return nil, errors.New("KAFKA_TOPIC environment variable is not set")
	}
	if cfg.KafkaGroupID == "" {
		return nil, errors.New("KAFKA_GROUP_ID environment variable is not set")
	}
	log.Println("Configuration loaded successfully.")
	return cfg, nil
}

type App struct {
	Config   *Config
	DB       *gorm.DB
	Cache    *cache.OrderCache
	Service  service.OrderService
	Consumer *kafka.Consumer
	Server   *http.Server
}

func NewApp(cfg *Config) (*App, error) {
	database, err := db.InitDB(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("could not initialize database: %w", err)
	}

	orderCache := cache.NewOrderCache()

	orderRepo := repository.NewOrderRepository(database)
	orderService := service.NewOrderService(orderRepo, orderCache)
	orderHandler := handler.NewOrderHandler(orderService)

	if err := orderService.RestoreCache(); err != nil {
		return nil, fmt.Errorf("failed to restore cache: %w", err)
	}

	kafkaConsumer, err := kafka.InitKafkaConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, orderService)
	if err != nil {
		return nil, fmt.Errorf("could not initialize Kafka consumer: %w", err)
	}

	router := setupRouter(orderHandler)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Config:   cfg,
		DB:       database,
		Cache:    orderCache,
		Service:  orderService,
		Consumer: kafkaConsumer,
		Server:   server,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	log.Println("Starting application...")
	go a.Consumer.Run(ctx)
	go func() {
		log.Printf("Server starting and listening on port %s", a.Config.Port)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func (a *App) Shutdown() {
	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := a.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
	}
}

func setupRouter(orderHandler *handler.OrderHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Get("/order/{id}", orderHandler.GetOrderByIDHandler)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	r.Handle("/*", http.FileServer(http.Dir("web")))
	return r
}

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("Application initialization failed: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app.Run(ctx)
	<-ctx.Done()
	app.Shutdown()
	log.Println("Application shut down gracefully.")
}
