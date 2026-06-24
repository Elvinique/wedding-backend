package main

import (
	"log"
	"os"
	"wedding-backend/config"
	"wedding-backend/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Connect to database
	config.ConnectDatabase()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Wedding API v1.0",
	})

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://www.faithwedsjoe2026.com.ng,https://faithwedsjoe2026.com.ng,https://wedding-backend-production-aefb.up.railway.app",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Rate limiting middleware
	app.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 60,
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "wedding-api"})
	})

	// RSVP routes
	app.Post("/api/rsvp", handlers.SubmitRSVP)
	app.Get("/api/rsvp/verify/:token", handlers.VerifyQR)

	// Admin routes
	app.Get("/api/admin/rsvps", handlers.GetAllRSVPs)

	// Guestbook routes
	app.Get("/api/guestbook", handlers.GetMessages)
	app.Post("/api/guestbook", handlers.PostMessage)

	// Serve QR images
	app.Static("/api/qr", "./qr_images")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
