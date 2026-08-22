package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"store_auth/internal/auth"
	"store_auth/internal/config"
	"store_auth/internal/jwt"
	"store_auth/internal/middleware"
	"store_auth/internal/otp"
	"store_auth/internal/platform/kafka"
	"store_auth/internal/platform/redis"
	"store_auth/internal/router"
	"store_auth/internal/user"
	"store_auth/prisma/db"
)

func main() {
	log.Println("[SERVER SETUP] Starting Auth Microservice (Feature-Based Architecture)...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load environment configuration: %v", err)
	}

	prismaClient := db.NewClient()
	if err := prismaClient.Connect(); err != nil {
		log.Fatalf("[FATAL] Failed to connect to database: %v", err)
	}
	defer func() {
		if err := prismaClient.Disconnect(); err != nil {
			log.Printf("[WARNING] Error disconnecting database client: %v", err)
		}
	}()
	log.Println("[SERVER SETUP] Connected to database successfully.")

	rdb, err := redis.NewClient(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Printf("[WARNING] Redis client initialization warning: %v", err)
	}

	// Initialize Kafka producer for domain event publishing
	var kafkaProducer kafka.Producer
	if strings.ToLower(cfg.OTPProvider) == "kafka" {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}
		producer := kafka.NewProducer(brokers)
		defer func() {
			if err := producer.Close(); err != nil {
				log.Printf("[WARNING] Error closing Kafka producer: %v", err)
			}
		}()
		kafkaProducer = producer
		log.Printf("[SERVER SETUP] Initialized Kafka producer for brokers: %v", brokers)
	}

	// Initialize repositories per feature
	userRepo := auth.NewRepository(prismaClient)
	refreshRepo := auth.NewRefreshRepository(prismaClient)
	otpRepo := otp.NewRepository(prismaClient)

	// Initialize services per feature
	jwtSvc, err := jwt.NewService(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath, cfg.JWTAccessExpiryMinutes)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize JWT service: %v", err)
	}

	otpSender := otp.NewOTPSender(cfg.OTPProvider, kafkaProducer, cfg.KafkaTopicAuthEvents)
	otpSvc := otp.NewService(otpRepo, otpSender, cfg.OTPExpiryMinutes, cfg.OTPMaxAttempts)
	authSvc := auth.NewService(userRepo, refreshRepo, otpSvc, jwtSvc, rdb, cfg.BcryptCost, cfg.JWTRefreshExpiryDays)

	// Initialize and run Kafka user lifecycle consumer in background
	var userEventsConsumer *user.Consumer
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()
	var consumerWg sync.WaitGroup

	if cfg.EnableUserEventsConsumer && cfg.KafkaBrokers != "" {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}
		kafkaReader := kafka.NewConsumer(brokers, cfg.KafkaTopicUserEvents, cfg.KafkaConsumerGroupAuth)
		userEventsConsumer = user.NewConsumer(kafkaReader, userRepo, refreshRepo, rdb)

		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			userEventsConsumer.Start(consumerCtx)
		}()
		log.Printf("[SERVER SETUP] Started Kafka user lifecycle consumer on topic '%s' (group: '%s')",
			cfg.KafkaTopicUserEvents, cfg.KafkaConsumerGroupAuth)
	}

	// Initialize handlers per feature
	isProd := cfg.Env == "production"
	authHandler := auth.NewHandler(authSvc, isProd)
	jwksHandler := jwt.NewHandler(jwtSvc)
	rateLimiter := middleware.NewRateLimiter(rdb, cfg.RateLimitMaxRequests, cfg.RateLimitWindowSec)

	// Wire router
	r := router.SetupRouter(authHandler, jwksHandler, rateLimiter, jwtSvc, rdb)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[SERVER ONLINE] Auth service listening on http://localhost%s", addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("[SHUTDOWN] Received signal %v. Initiating graceful shutdown...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cancelConsumer()
		if userEventsConsumer != nil {
			if err := userEventsConsumer.Close(); err != nil {
				log.Printf("[WARNING] Error closing Kafka user consumer: %v", err)
			}
		}
		consumerWg.Wait()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[WARNING] Server forced to shutdown: %v", err)
		}
		log.Println("[SHUTDOWN] Server exiting cleanly.")
	}
}
