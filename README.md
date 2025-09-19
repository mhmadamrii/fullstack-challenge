## How to run the stack locally

To run the stack locally, you will need to have Docker and Docker Compose installed.

1. Clone the repository:

```bash
git clone <repository-url>
```

2. Navigate to the project directory:

```bash
cd fullstack-challenge
```

3. Run the stack:

```bash
docker-compose up
```

This will start the following services:

*   `product-service`: A NestJS service for managing products.
*   `order-service`: A Go service for managing orders.
*   `postgres`: A PostgreSQL database for storing data.
*   `redis`: a Redis instance for caching.
*   `rabbitmq`: A RabbitMQ instance for messaging.

## Architecture

This project is a microservices-based application that consists of two services:

*   **`product-service`**: A NestJS service that is responsible for managing products. It exposes a REST API for creating, retrieving, updating, and deleting products. It also uses Redis for caching and RabbitMQ for asynchronous communication with the `order-service`.
*   **`order-service`**: A Go service that is responsible for managing orders. It exposes a REST API for creating and retrieving orders. It communicates with the `product-service` via RabbitMQ to update the product stock when an order is created.

The services are containerized using Docker and can be run locally using Docker Compose.

## API Examples

### Product Service

*   **Get all products:**

```bash
curl http://localhost:3000/products
```

*   **Create a new product:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name": "Product 1", "price": 10.99, "qty": 100}' http://localhost:3000/products
```

### Order Service

*   **Get all orders:**

```bash
curl http://localhost:8080/orders
```

*   **Create a new order:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"productId": "<product-id>"}' http://localhost:8080/orders
```
