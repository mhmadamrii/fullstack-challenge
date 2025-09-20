## How to run the stack locally

To run the stack locally, you will need to have Docker and Docker Compose installed.

1. Clone the repository:

```bash
git clone https://github.com/mhmadamrii/fullstack-challenge
```

2. Navigate to the project directory:

```bash
cd fullstack-challenge
```

3. Run the stack:

```bash
docker-compose up --build
```

This will spin up the following services:

- `product-service`: A NestJS service for managing products.
- `order-service`: A Go service for managing orders.
- `postgres`: A PostgreSQL database for storing data.
- `redis`: a Redis instance for caching.
- `rabbitmq`: A RabbitMQ instance for messaging.

## Architecture

This project is a microservices-based application that consists of two services:

- **`product-service`**: A NestJS service that is responsible for managing products. It exposes a REST API for creating, retrieving, updating, and deleting products. It also uses Redis for caching and RabbitMQ for asynchronous communication with the `order-service`.
- **`order-service`**: A Go service that is responsible for managing orders. It exposes a REST API for creating and retrieving orders. It communicates with the `product-service` via RabbitMQ to update the product stock when an order is created.

## Available Endpoints

- `POST /products` → create product **(product service)**
- `GET /products/:id` → fetch product cached using redis **(product service)**
- `GET /orders` → get all orders **(product service)**
- `POST /orders` → create order for existing product **(order service)**
- `GET /orders/product/:productId` → fetch product’s orders (cached) **(order service)**
- `DELETE /orders` → delete all created orders **(order service)**

## API Examples

### Product Service

- **Get all products:**

```bash
curl http://localhost:3000/products
```

- **Create a new product:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name": "Macbook Pro M3", "price": 2000, "qty": 100}' http://localhost:3000/products
```

- **Get product by id:**

```bash
curl http://localhost:3000/products/<product-id>
```

### Order Service

- **Get all orders:**

```bash
curl http://localhost:8080/orders
```

- **Create a new order:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"productId": "<product-id>"}' http://localhost:8080/orders
```

- **Get orders by product id:**

```bash
curl http://localhost:8080/orders/product/<product-id>
```

## Testing with k6

Creating order took around ~900RPS

![alt text](https://oyluendsrr.ufs.sh/f/heCK4TZGuZCFqIcXv2r8hsLqQr6XRyHSANKdDvj1I2YlfPoi)

Before running the test, make sure you have:

- k6 installed locally.
- A product available in the `products` table.
- Increased the product quantity to a high number, for instance, 50000.

## How to run the tests locally?

1. Install k6:

```bash
brew install k6
```

2. Run the tests:

```bash
k6 run ./__test__/k6/load-test.js
```
