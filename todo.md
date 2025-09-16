Big Picture Flow

We want:

Create product → store in DB → cache on read → publish event.

Create order → validate product via product-service → store in DB → publish event.

Consume order events → product-service reduces stock → order-service logs/async tasks.

Redis + RabbitMQ glue both services.

Docker Compose ties everything up.

Unit tests + load tests.

# Product Service

## Step-by-Step Implementation Flow

Phase 1: Product-service (NestJS)

- Set up DB connection (Postgres) + Product entity (id, name, price, qty).

- Implement POST /products → insert into DB → return new product.

- Implement GET /products/:id → check Redis first → fallback to DB → cache result.

- Add RabbitMQ publisher (e.g., publish product.created when product is created).

- Add RabbitMQ consumer for order.created → reduce product qty.Step-by-Step Implementation Flow
