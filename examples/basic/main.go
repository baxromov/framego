package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/baxromov/framego/pkg/api"
	"github.com/baxromov/framego/pkg/config"
	"github.com/baxromov/framego/pkg/graphql"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/models"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
	"github.com/baxromov/framego/pkg/serializer"
)

// User represents a user model
type User struct {
	models.Model
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	// Load configuration from file
	cfg, err := config.LoadFromFile("examples/basic/config.json")
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

	// Create a new user model
	userModel := models.NewModel("users")
	userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	userModel.AddField("email", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100), models.WithUnique())
	userModel.AddField("password", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	userModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Create a new ORM instance using configuration
	orm, err := orm.New(cfg.ToORMConfig())
	if err != nil {
		log.Fatalf("Failed to create ORM: %v", err)
	}
	defer orm.Close()

	// Register the user model with the ORM
	if err := orm.RegisterModel(userModel); err != nil {
		log.Fatalf("Failed to register model: %v", err)
	}

	// Create a new router
	r := router.New()

	// Create GraphQL handler if enabled
	var graphqlHandler *graphql.Handler
	if cfg.GraphQL.Enabled {
		graphqlHandler = graphql.New(orm)
		log.Println("GraphQL support enabled")
	}

	// Setup product API
	setupProductAPI(orm, r, graphqlHandler)

	// Setup order API
	setupOrderAPI(orm, r, graphqlHandler)

	// Create tables
	if err := orm.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// Create a new API controller for the user model
	userController := api.NewController(orm, userModel, "/api/users")

	// Create a custom serializer for the user model
	userSerializer := serializer.New(userModel)
	// Make password write-only
	userSerializer.AddField("password", reflect.TypeOf(""), serializer.WithWriteOnly())
	// Add email validator
	userSerializer.AddField("email", reflect.TypeOf(""), serializer.WithValidator(serializer.EmailValidator()))

	// Set the custom serializer for the controller
	userController.SetSerializer(userSerializer)

	// Register routes
	apiGroup := r.Group("/api")

	// Public routes
	apiGroup.GET("/users", userController.List)
	apiGroup.GET("/users/:id", userController.Get)

	// Protected routes
	apiGroup.POST("/users", userController.Create, middleware.Auth)
	apiGroup.PUT("/users/:id", userController.Update, middleware.Auth)
	apiGroup.DELETE("/users/:id", userController.Delete, middleware.Auth)

	// Register user model with GraphQL if enabled
	if cfg.GraphQL.Enabled && graphqlHandler != nil {
		// Register user model with GraphQL
		if err := graphqlHandler.RegisterModel(userModel); err != nil {
			log.Printf("Failed to register user model with GraphQL: %v", err)
		} else {
			fmt.Println("User model registered with GraphQL")
		}

		// Register GraphQL handler
		http.Handle(cfg.GraphQL.Path, graphqlHandler)

		log.Printf("GraphQL endpoint available at http://%s:%d%s", cfg.Server.Host, cfg.Server.Port, cfg.GraphQL.Path)
	}

	// Start the server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started at http://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
