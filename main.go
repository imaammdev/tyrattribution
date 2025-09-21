package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tyrattribution/config"
	"tyrattribution/consumer"
	"tyrattribution/database"
	"tyrattribution/handler"
	"tyrattribution/publisher"
	"tyrattribution/redis"
	"tyrattribution/repository"
	"tyrattribution/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.OpenDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	clickEventRepo := repository.NewClickEventRepository(db)
	conversionEventRepo := repository.NewConversionEventRepository(db)
	campaignRepo := repository.NewCampaignRepository(db)
	campaignJournalRepo := repository.NewCampaignJournalRepository(db)

	clickEventService := service.NewClickEventService(clickEventRepo, redisClient)
	conversionEventService := service.NewConversionEventService(conversionEventRepo, clickEventService, redisClient, cfg)
	campaignJournalService := service.NewCampaignJournalService(campaignJournalRepo, campaignRepo, clickEventRepo, conversionEventRepo, redisClient, db)
	campaignStatisticsService := service.NewCampaignStatisticsService(campaignJournalRepo, redisClient, db)
	clickEventPublisher, err := publisher.NewClickEventPublisher()
	if err != nil {
		log.Fatalf("Failed to create click event publisher: %v", err)
	}

	conversionEventPublisher, err := publisher.NewConversionEventPublisher()
	if err != nil {
		log.Fatalf("Failed to create conversion event publisher: %v", err)
	}

	mux := handler.SetupRoutes(clickEventPublisher, conversionEventPublisher, campaignJournalService, campaignStatisticsService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.StartClickEventConsumer(ctx, clickEventService)
	go consumer.StartConversionEventConsumer(ctx, conversionEventService)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel context to stop consumer
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
