# Django-Style Example Project

This example project demonstrates how to structure a FrameGo application similar to a Django project. The project is organized into separate apps, each with its own models, serializers, and views.

## Project Structure

```
django_style/
├── cmd/
│   └── server/
│       └── main.go         # Main application entry point
├── config/
│   └── config.json         # Application configuration
└── internal/
    ├── users/              # Users app
    │   ├── models.go       # User model definitions
    │   ├── serializers.go  # User serializer definitions
    │   └── views.go        # User API routes and controllers
    ├── products/           # Products app
    │   ├── models.go       # Product model definitions
    │   ├── serializers.go  # Product serializer definitions
    │   └── views.go        # Product API routes and controllers
    └── orders/             # Orders app
        ├── models.go       # Order and OrderItem model definitions
        ├── serializers.go  # Order and OrderItem serializer definitions
        └── views.go        # Order API routes and controllers
```

## Apps

### Users App

The Users app handles user authentication and profile management. It includes:

- User model with fields for username, email, and password
- User serializer with password write-only protection and email validation
- API endpoints for listing, retrieving, creating, updating, and deleting users

### Products App

The Products app manages the product catalog. It includes:

- Product model with fields for name, description, price, and stock
- Product serializer with price and stock validation
- API endpoints for listing, retrieving, creating, updating, and deleting products

### Orders App

The Orders app handles order processing and tracking. It includes:

- Order model with fields for user, total price, and status
- OrderItem model with fields for order, product, quantity, and price
- Order serializer with status validation
- OrderItem serializer with quantity validation
- API endpoints for listing, retrieving, creating, updating, and deleting orders and order items

## Running the Example

To run the example, use the following command from the project root:

```bash
go run examples/django_style/cmd/server/main.go
```

This will start the server at http://localhost:8080 with the following endpoints:

- User API: http://localhost:8080/api/users
- Product API: http://localhost:8080/api/products
- Order API: http://localhost:8080/api/orders
- Order Item API: http://localhost:8080/api/order-items
- GraphQL API: http://localhost:8080/graphql (if enabled in config)

## API Endpoints

### Users

- GET /api/users - List all users
- GET /api/users/:id - Get a specific user
- POST /api/users - Create a new user
- PUT /api/users/:id - Update a user
- DELETE /api/users/:id - Delete a user

### Products

- GET /api/products - List all products
- GET /api/products/:id - Get a specific product
- POST /api/products - Create a new product
- PUT /api/products/:id - Update a product
- DELETE /api/products/:id - Delete a product

### Orders

- GET /api/orders - List all orders
- GET /api/orders/:id - Get a specific order
- POST /api/orders - Create a new order
- PUT /api/orders/:id - Update an order
- DELETE /api/orders/:id - Delete an order

### Order Items

- GET /api/order-items - List all order items
- GET /api/order-items/:id - Get a specific order item
- POST /api/order-items - Create a new order item
- PUT /api/order-items/:id - Update an order item
- DELETE /api/order-items/:id - Delete an order item

## GraphQL

If GraphQL is enabled in the configuration, you can use the GraphQL API at http://localhost:8080/graphql. The GraphQL API provides the same functionality as the REST API, but with the flexibility of GraphQL.

Example GraphQL queries (these are illustrative examples and may need to be adjusted based on your actual GraphQL schema):

```
# Get all users
query {
  users {
    id
    username
    email
  }
}

# Get a specific product
query {
  product(id: 1) {
    id
    name
    description
    price
    stock
  }
}

# Create a new order
mutation {
  createOrder(user_id: 1, total_price: 100.0, status: "pending") {
    id
    user_id
    total_price
    status
  }
}
```
