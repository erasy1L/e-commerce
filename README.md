# E-commerce

E-commerce microservices project

## Features

- Creating users, products
- Placing an order and making payments with idempotency check
- gRPC communication between microservices

## Installation & Usage

1. Navigate to the project directory: `cd ecommerce`.
2. Rename .env.example to .env and change variables accordingly.
3. Start the docker containers: `make up`.
4. Navigate to swagger docs at http://localhost:8080/swagger/index.html.

## Libraries

1. [go-chi](https://github.com/go-chi/chi) as router
2. [zerolog](https://github.com/rs/zerolog) as logger
