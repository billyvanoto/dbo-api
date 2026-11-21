package main

import (
	"context"
	"dbo-api/config"
	"dbo-api/routes"
	"dbo-api/utils/validation"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	config.ConnectDatabase()
	validation.RegisterCustomValidations()

	r := routes.SetupRouter()

	// Healthcheck endpoint for Docker / load balancers
	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := config.DB.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"db":     "disconnected",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Listen for SIGINT (Ctrl+C) and SIGTERM (docker stop / docker compose down)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until shutdown signal is received
	<-ctx.Done()
	log.Println("Shutdown signal received, draining active connections...")

	// Allow up to 10 seconds for in-flight requests to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	// Close underlying database connection pool
	if sqlDB, err := config.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	log.Println("Server exited cleanly")
}
