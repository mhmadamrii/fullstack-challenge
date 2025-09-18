package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/mhmadamrii/order-service/internal/http"
	"github.com/mhmadamrii/order-service/internal/orders"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	orderService := orders.NewService()
	orderHandlers := http.NewHandlers(orderService)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Order Service is running 🚀")
	})

	app.Post("/orders", orderHandlers.CreateOrder)
	app.Get("/orders/:productId", orderHandlers.GetOrdersByProductID)
	app.Get("/products", orderHandlers.GetAllProducts)

	log.Printf("Server listening on port %s", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
