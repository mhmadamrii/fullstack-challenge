package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mhmadamrii/order-service/internal/models"
)



type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Service struct {
	db                *gorm.DB
	httpClient        *http.Client
	redisClient       *redis.Client
	rabbitmqChannel   *amqp091.Channel
	productServiceURL string
}

func NewService() *Service {
	// DB
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	err = db.AutoMigrate(&models.Order{})
	if err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migration complete")

	// Redis
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}
	log.Println("✅ Redis connected successfully")
	rdb := redis.NewClient(opt)

	// RabbitMQ
	conn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	productServiceURL := os.Getenv("PRODUCT_SERVICE")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:3000"
	}

	return &Service{
		db:                db,
		httpClient:        &http.Client{},
		redisClient:       rdb,
		rabbitmqChannel:   ch,
		productServiceURL: productServiceURL,
	}
}

func (s *Service) CreateOrder(productID string, quantity int) (*models.Order, error) {
	resp, err := s.httpClient.Get(s.productServiceURL + "/products/" + productID)
	if err != nil {
		return nil, errors.New("failed to contact product-service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("product not found")
	}

	var product struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, err
	}

	order := &models.Order{
		ID:         uuid.New().String(),
		ProductID:  productID,
		TotalPrice: product.Price * float64(quantity),
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}
	if err := s.db.Create(order).Error; err != nil {
		return nil, err
	}

	if err := s.publishOrderCreatedV2(order); err != nil {
		log.Printf("Failed to publish order.created: %s", err)
	}

	return order, nil
}

func (s *Service) publishOrderCreated(order *models.Order) error {
	fmt.Println("Creating order.created event")
	q, err := s.rabbitmqChannel.QueueDeclare(
		"order.created", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return s.rabbitmqChannel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (s *Service) publishOrderCreatedV2(order *models.Order) error {
	fmt.Println("Attempting to publish order.created event")
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return s.rabbitmqChannel.Publish(
		"events",        // exchange
		"order.created", // routing key
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (s *Service) GetOrdersByProductID(productID string) ([]*models.Order, error) {
	cacheKey := "orders:product:" + productID
	ctx := context.Background()

	cachedOrders, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var orders []*models.Order
		if err := json.Unmarshal([]byte(cachedOrders), &orders); err == nil {
			log.Println("✅ Cache hit for orders:product:" + productID)
			return orders, nil
		}
	}

	log.Println("🗑️ Cache miss for orders:product:" + productID)
	var orders []*models.Order
	if err := s.db.Where("\"productId\" = ?", productID).Find(&orders).Error; err != nil {
		return nil, err
	}

	// Cache the result
	ordersJSON, err := json.Marshal(orders)
	if err == nil {
		s.redisClient.Set(ctx, cacheKey, ordersJSON, 10*time.Minute)
	}

	return orders, nil
}

func (s *Service) GetAllProducts() ([]*Product, error) {
	cacheKey := "products:all"
	ctx := context.Background()

	cachedProducts, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []*Product
		if err := json.Unmarshal([]byte(cachedProducts), &products); err == nil {
			log.Println("✅ Cache hit for " + cacheKey)
			return products, nil
		}
	}

	log.Println("🗑️ Cache miss for " + cacheKey)
	resp, err := s.httpClient.Get(s.productServiceURL + "/products")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch products from product-service")
	}

	var products []*Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}

	productsJSON, err := json.Marshal(products)
	if err == nil {
		s.redisClient.Set(ctx, cacheKey, productsJSON, 10*time.Minute)
	}

	return products, nil
}
