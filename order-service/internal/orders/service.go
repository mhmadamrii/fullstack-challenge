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

type OrderJob struct {
	Order *models.Order
}

type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Qty       int       `json:"qty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Service struct {
	db                *gorm.DB
	httpClient        *http.Client
	redisClient       *redis.Client
	rabbitmqChannel   *amqp091.Channel
	productServiceURL string
	jobQueue          chan *OrderJob
	workers           int
}

func NewService() *Service {
	// DB
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect database: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get generic DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	log.Println("✅ DB connection pool configured")

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
	rdb := redis.NewClient(opt)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// RabbitMQ
	conn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}
	log.Println("✅ Connected to RabbitMQ")

	productServiceURL := os.Getenv("PRODUCT_SERVICE")
	if productServiceURL == "" {
		productServiceURL = "http://localhost:3000"
	}

	s := &Service{
		db:                db,
		redisClient:       rdb,
		rabbitmqChannel:   ch,
		productServiceURL: productServiceURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        1000,
				MaxIdleConnsPerHost: 500,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: 2 * time.Second,
		},
	}

	s.jobQueue = make(chan *OrderJob, 5000)
	s.workers = 8
	for i := 0; i < s.workers; i++ {
		go s.startWorker(i)
	}

	return s
}

func (s *Service) CreateOrder(productID string) (*models.Order, error) {
	ctx := context.Background()
	cacheKey := "product:" + productID
	cacheKeyOrdersByProductID := "orders:product:" + productID

	var product struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Qty   int     `json:"qty"`
	}

	if cached, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			log.Printf("✅ Local cache hit for product %s", productID)
		}
	}

	if product.ID == "" {
		resp, err := s.httpClient.Get(s.productServiceURL + "/products/" + productID)
		if err != nil {
			return nil, errors.New("failed to contact product-service")
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("product not found")
		}

		if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
			return nil, err
		}

		if jsonData, err := json.Marshal(product); err == nil {
			_ = s.redisClient.Set(ctx, cacheKey, jsonData, 60*time.Second).Err()
		}
	}

	if product.Qty <= 0 {
		return nil, errors.New("product is out of stock")
	}

	order := &models.Order{
		ID:         uuid.New().String(),
		ProductID:  productID,
		TotalPrice: product.Price,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	if err := s.redisClient.Del(ctx, cacheKeyOrdersByProductID).Err(); err != nil {
		log.Printf("⚠️ Failed to invalidate orders cache for product %s: %v", productID, err)
	} else {
		log.Printf("🗑️ Invalidated orders cache for product %s", productID)
	}

	select {
	case s.jobQueue <- &OrderJob{Order: order}:
		return order, nil
	default:
		return nil, errors.New("server overloaded, try again")
	}
}

func (s *Service) publishOrderCreatedV2(order *models.Order) error {
	body, err := json.Marshal(order)
	if err != nil {
		return err
	}
	fmt.Println("Attempting to publish order.created event")

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
	ctx := context.Background()
	cacheKey := "orders:product:" + productID

	if cachedOrders, err := s.redisClient.Get(ctx, cacheKey).Result(); err == nil && cachedOrders != "" {
		var orders []*models.Order
		if err := json.Unmarshal([]byte(cachedOrders), &orders); err == nil {
			log.Printf("✅ Cache hit for orders of product %s", productID)
			return orders, nil
		}
		log.Printf("⚠️ Failed to unmarshal cached orders for product %s: %v", productID, err)
	} else {
		log.Printf("🗑️ Cache miss for orders of product %s", productID)
	}

	var orders []*models.Order
	if err := s.db.Where("\"productId\" = ?", productID).Find(&orders).Error; err != nil {
		return nil, err
	}

	if jsonData, err := json.Marshal(orders); err == nil {
		if err := s.redisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err(); err != nil {
			log.Printf("⚠️ Failed to cache orders for product %s: %v", productID, err)
		} else {
			log.Printf("✅ Cached orders for product %s", productID)
		}
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

func (s *Service) startWorker(id int) {
	log.Printf("order-worker %d started", id)
	for job := range s.jobQueue {
		if job == nil || job.Order == nil {
			continue
		}
		ctx := context.Background()

		// Persist order (simple retry loop)
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			err = s.db.Create(job.Order).Error
			if err == nil {
				break
			}
			log.Printf("worker %d: failed to persist order %s (attempt %d): %v", id, job.Order.ID, attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		if err != nil {
			log.Printf("worker %d: permanent failure persisting order %s: %v", id, job.Order.ID, err)
			continue
		}

		cacheKey := "orders:product:" + job.Order.ProductID
		if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("worker %d: failed to invalidate orders cache %s: %v", id, cacheKey, err)
		}

		if err := s.publishOrderCreatedV2(job.Order); err != nil {
			log.Printf("worker %d: failed to publish order.created for %s: %v", id, job.Order.ID, err)
		}
	}
}
