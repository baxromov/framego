package main

import (
	"fmt"
	"log"
	"net/http"

	// Import database drivers
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver


	"github.com/baxromov/framego/examples/django_style/internal/orders"
	"github.com/baxromov/framego/examples/django_style/internal/products"
	"github.com/baxromov/framego/examples/django_style/internal/users"
	"github.com/baxromov/framego/pkg/config"
	"github.com/baxromov/framego/pkg/graphql"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
)

func main() {
	// Load configuration from file
	cfg, err := config.LoadFromFile("examples/django_style/config/config.json")
	if err != nil {
		log.Printf("Failed to load config from file: %v", err)
		log.Println("Using default configuration")
		cfg = config.DefaultConfig()
	}

	// Print configuration in debug mode
	if cfg.Debug {
		log.Println("Debug mode enabled")
		log.Printf("Configuration: %+v", cfg)
	}

	// Create a new ORM instance using configuration
	orm, err := orm.New(cfg.ToORMConfig())
	if err != nil {
		log.Fatalf("Failed to create ORM: %v", err)
	}
	defer orm.Close()

	// Create a new router
	r := router.New()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// Create GraphQL handler if enabled
	var graphqlHandler *graphql.Handler
	if cfg.GraphQL.Enabled {
		graphqlHandler = graphql.New(orm)
		log.Println("GraphQL support enabled")
	}

	// Setup user API
	users.SetupUserAPI(orm, r)

	// Setup product API
	products.SetupProductAPI(orm, r)

	// Setup order API
	orders.SetupOrderAPI(orm, r)

	// Create tables
	if err := orm.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Register GraphQL handler if enabled
	if cfg.GraphQL.Enabled && graphqlHandler != nil {
		// Register models with GraphQL
		userModel := users.CreateUserModel()
		productModel := products.CreateProductModel()
		orderModel := orders.CreateOrderModel()
		orderItemModel := orders.CreateOrderItemModel()

		if err := graphqlHandler.RegisterModel(userModel); err != nil {
			log.Printf("Failed to register user model with GraphQL: %v", err)
		} else {
			fmt.Println("User model registered with GraphQL")
		}

		if err := graphqlHandler.RegisterModel(productModel); err != nil {
			log.Printf("Failed to register product model with GraphQL: %v", err)
		} else {
			fmt.Println("Product model registered with GraphQL")
		}

		if err := graphqlHandler.RegisterModel(orderModel); err != nil {
			log.Printf("Failed to register order model with GraphQL: %v", err)
		} else {
			fmt.Println("Order model registered with GraphQL")
		}

		if err := graphqlHandler.RegisterModel(orderItemModel); err != nil {
			log.Printf("Failed to register order item model with GraphQL: %v", err)
		} else {
			fmt.Println("Order item model registered with GraphQL")
		}

		// Register GraphQL handler with the router for both POST and GET requests
		r.Handle("POST", cfg.GraphQL.Path, http.HandlerFunc(graphqlHandler.ServeHTTP))
		r.Handle("GET", cfg.GraphQL.Path, http.HandlerFunc(graphqlHandler.ServeHTTP))

		log.Printf("GraphQL endpoint available at http://%s:%d%s", cfg.Server.Host, cfg.Server.Port, cfg.GraphQL.Path)
	}

	// Start the server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started at http://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
