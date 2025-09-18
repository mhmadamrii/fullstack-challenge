package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhmadamrii/order-service/internal/orders"
)

type Handlers struct {
	orderService *orders.Service
}

func NewHandlers(orderService *orders.Service) *Handlers {
	return &Handlers{
		orderService: orderService,
	}
}

type createOrderRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func (h *Handlers) CreateOrder(c *fiber.Ctx) error {
	var req createOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	order, err := h.orderService.CreateOrder(req.ProductID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (h *Handlers) GetOrdersByProductID(c *fiber.Ctx) error {
	productID := c.Params("productId")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "productId is required",
		})
	}

	orders, err := h.orderService.GetOrdersByProductID(productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(orders)
}

func (h *Handlers) GetAllProducts(c *fiber.Ctx) error {
	products, err := h.orderService.GetAllProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(products)
}
