# E-commerce

E-commerce microservices project

## Features

- Creating users, products
- Placing an order and making payments with idempotency check
- gRPC communication between microservices

## Installation & Usage

```
gateway-1  | [00] Starting service
product-1  | [00] Starting service
order-1    | [00] Starting service
user-1     | [00] Starting service
payment-1  | [00] Starting service
migrate-1  | no change
migrate-1 exited with code 0
user-1     | [00] Service started on port 8081
product-1  | [00] Service started on port 8082
order-1    | [00] Service started on port 8083
payment-1  | [00] Service started on port 8084
gateway-1  | [00] Mounting service  /users on http://user:8081
gateway-1  | [00] Mounting service  /products on http://product:8082
gateway-1  | [00] Mounting service  /orders on http://order:8083
gateway-1  | [00] Mounting service  /payments on http://payment:8084
gateway-1  | [00] API Gateway started on port 8080 swagger on http://localhost:8080/swagger/index.html
```

1. Navigate to the project directory: `cd ecommerce`.
2. Rename .env.example to .env and change variables accordingly.
3. Start the docker containers: `make up`.
4. Navigate to swagger docs at http://localhost:8080/swagger/index.html.

## Libraries

1. [go-chi](https://github.com/go-chi/chi) as router
2. [zerolog](https://github.com/rs/zerolog) as logger

## TODO
1. add logging
2. add oauth
