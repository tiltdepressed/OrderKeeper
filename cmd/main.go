// @title OrderKeeper API
// @version 1.0
// @description API for managing orders
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"errors"
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
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"
)

type App struct {
	DB            *gorm.DB
	Cache         *cache.OrderCache
	Service       service.OrderService
	Handler       *handler.OrderHandler
	KafkaConsumer *kafka.Consumer
	Server        *http.Server
}

func initConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func initServices(db *gorm.DB, cache *cache.OrderCache) (service.OrderService, *handler.OrderHandler, error) {
	orderRepository := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepository, cache)
	orderHandler := handler.NewOrderHandler(orderService)

	return orderService, orderHandler, nil
}

func setupRouter(orderHandler *handler.OrderHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, ".js") {
				w.Header().Set("Content-Type", "application/javascript")
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Post("/order", orderHandler.CreateOrderHandler)
	r.Get("/order/{id}", orderHandler.GetOrderByIDHandler)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	r.Handle("/web/*", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	return r
}

func startServer(router *chi.Mux) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server started and listening on port %s", port)
		log.Printf("Web interface: http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return server
}

func gracefulShutdown(server *http.Server, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %s, shutting down...", sig)

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

func main() {
	initConfig()

	db, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	orderCache := cache.NewOrderCache()

	orderService, orderHandler, err := initServices(db, orderCache)
	if err != nil {
		log.Fatalf("Could not initialize services: %v", err)
	}

	if err := orderService.RestoreCache(); err != nil {
		log.Fatalf("Failed to restore cache: %v", err)
	}
	log.Printf("Cache restored with %d orders", orderCache.Count())

	kafkaConsumer, err := kafka.InitKafkaConsumer(orderService)
	if err != nil {
		log.Fatalf("Could not initialize Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Printf("Starting Kafka consumer (brokers: %s, topic: %s, group: %s)",
			os.Getenv("KAFKA_BROKERS"), os.Getenv("KAFKA_TOPIC"), os.Getenv("KAFKA_GROUP_ID"))
		kafkaConsumer.Run(ctx)
	}()

	router := setupRouter(orderHandler)
	server := startServer(router)

	gracefulShutdown(server, cancel)
}
