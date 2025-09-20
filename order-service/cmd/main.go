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

	app := fiber.New(fiber.Config{
		Prefork: false,
		AppName: "Order Service",
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Order Service is running 🚀")
	})

	app.Post("/orders-mock", orderHandlers.CreateOrderMock)
	app.Post("/orders", orderHandlers.CreateOrder)
	app.Get("/orders/product/:productId", orderHandlers.GetOrdersByProductID)
	app.Get("/products", orderHandlers.GetAllProducts)
	app.Delete("/orders", orderHandlers.DeleteAllOrders)

	log.Printf("Server listening on port %s", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}
