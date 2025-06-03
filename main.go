package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/baxromov/framego/pkg/api"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/models"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
	"github.com/baxromov/framego/pkg/serializer"

	// Import database drivers
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	
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
	// Create a new user model
	userModel := models.NewModel("users")
	userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	userModel.AddField("email", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100), models.WithUnique())
	userModel.AddField("password", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	userModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Create a new ORM instance
	ormConfig := orm.Config{
		Driver:   "sqlite3",
		Database: "test.db",
	}

	orm, err := orm.New(ormConfig)
	if err != nil {
		log.Fatalf("Failed to create ORM: %v", err)
	}
	defer orm.Close()

	// Register the user model with the ORM
	if err := orm.RegisterModel(userModel); err != nil {
		log.Fatalf("Failed to register model: %v", err)
	}

	// Create tables
	if err := orm.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create a new router
	r := router.New()

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

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
