package main

import (
	"file-management-service/config"
	"file-management-service/routes"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Global variable to hold the configuration
var AppConfig *config.Config

func main() {
	e := echo.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.SetOutput(os.Stderr)
	// Apply rate limiter middleware
	rateLimiterConfig := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 30, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}

	e.Use(middleware.RateLimiterWithConfig(rateLimiterConfig))

	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	// Assign the configuration to the global variable
	AppConfig = config

	// Register routes
	routes.RegisterRoutes(e, AppConfig)

	// Start the server
	e.Start(":8080")
	log.Println("Server Started!!!")
}
