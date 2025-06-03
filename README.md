# FrameGo

FrameGo is a lightweight web framework for Go, inspired by Django's REST framework. It provides a simple and intuitive way to build REST APIs with Go.

## Features

- **Models**: Define your data models with field types, constraints, and relationships
- **ORM**: Interact with your database using a simple and powerful ORM
- **API**: Create REST endpoints with minimal code
- **Serializers**: Transform data between your models and JSON
- **Router**: Handle HTTP requests with a flexible router
- **Middleware**: Add functionality to your API with middleware
- **Configuration**: Manage application settings with support for different databases and environments
- **GraphQL**: Expose your models via GraphQL in addition to REST

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Creating Models](#creating-models)
  - [Field Types](#field-types)
  - [Field Constraints](#field-constraints)
  - [Model Relationships](#model-relationships)
- [Working with ORM](#working-with-orm)
  - [Connecting to Databases](#connecting-to-databases)
  - [CRUD Operations](#crud-operations)
  - [Custom Queries](#custom-queries)
- [Creating Serializers](#creating-serializers)
  - [Field Customization](#field-customization)
  - [Validation](#validation)
- [Creating Views and Routes](#creating-views-and-routes)
  - [Controllers](#controllers)
  - [Routing](#routing)
  - [Middleware](#middleware)
- [Configuration](#configuration)
  - [Loading Configuration](#loading-configuration)
  - [Database Configuration](#database-configuration)
- [Using GraphQL](#using-graphql)
  - [Setting Up GraphQL](#setting-up-graphql)
  - [Queries and Mutations](#queries-and-mutations)
- [Examples](#examples)
  - [Blog Application](#blog-application)
  - [E-commerce Application](#e-commerce-application)
- [License](#license)

## Installation

```bash
go get github.com/baxromov/framego
```

## Quick Start

```go
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
```

## Creating Models

Models in FrameGo represent your database tables. They define the structure of your data and the relationships between different entities.

### Field Types

FrameGo supports various field types that map to database column types:

```go
// Integer types
userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey())
userModel.AddField("age", reflect.TypeOf(0))

// String types
userModel.AddField("name", reflect.TypeOf(""), models.WithMaxLength(100))
userModel.AddField("description", reflect.TypeOf(""))

// Boolean types
userModel.AddField("is_active", reflect.TypeOf(true))

// Float types
userModel.AddField("price", reflect.TypeOf(0.0))

// Time types
userModel.AddField("created_at", reflect.TypeOf(time.Time{}))
```

### Field Constraints

You can add various constraints to your fields:

```go
// Primary key
userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey())

// Auto-increment
userModel.AddField("id", reflect.TypeOf(0), models.WithAutoIncrement())

// Not null
userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull())

// Unique
userModel.AddField("email", reflect.TypeOf(""), models.WithUnique())

// Default value
userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithDefault(time.Now()))

// Maximum length (for strings)
userModel.AddField("username", reflect.TypeOf(""), models.WithMaxLength(50))
```

### Model Relationships

FrameGo supports various types of relationships between models:

#### One-to-Many Relationship

```go
// User model
userModel := models.NewModel("users")
userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50))

// Post model with a foreign key to User
postModel := models.NewModel("posts")
postModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
postModel.AddField("title", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
postModel.AddField("content", reflect.TypeOf(""), models.WithNotNull())
postModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
    models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
```

#### Many-to-Many Relationship

```go
// Product model
productModel := models.NewModel("products")
productModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
productModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
productModel.AddField("price", reflect.TypeOf(0.0), models.WithNotNull())

// Category model
categoryModel := models.NewModel("categories")
categoryModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
categoryModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50))

// ProductCategory join table
productCategoryModel := models.NewModel("product_categories")
productCategoryModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
productCategoryModel.AddField("product_id", reflect.TypeOf(0), models.WithNotNull(), 
    models.WithForeignKey("products", "id", "CASCADE", "CASCADE"))
productCategoryModel.AddField("category_id", reflect.TypeOf(0), models.WithNotNull(), 
    models.WithForeignKey("categories", "id", "CASCADE", "CASCADE"))
```

## Working with ORM

The ORM (Object-Relational Mapper) provides a simple way to interact with your database.

### Connecting to Databases

FrameGo supports various database drivers:

#### SQLite

```go
ormConfig := orm.Config{
    Driver:   "sqlite3",
    Database: "test.db",
}

orm, err := orm.New(ormConfig)
if err != nil {
    log.Fatalf("Failed to create ORM: %v", err)
}
defer orm.Close()
```

#### MySQL

```go
ormConfig := orm.Config{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    User:     "root",
    Password: "password",
    Database: "myapp",
}

orm, err := orm.New(ormConfig)
if err != nil {
    log.Fatalf("Failed to create ORM: %v", err)
}
defer orm.Close()
```

#### PostgreSQL

```go
ormConfig := orm.Config{
    Driver:   "postgres",
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "myapp",
}

orm, err := orm.New(ormConfig)
if err != nil {
    log.Fatalf("Failed to create ORM: %v", err)
}
defer orm.Close()
```

### CRUD Operations

#### Create

```go
// Create a new user
id, err := orm.Create("users", map[string]interface{}{
    "username": "john",
    "email":    "john@example.com",
    "password": "password123",
})
if err != nil {
    log.Fatalf("Failed to create user: %v", err)
}
fmt.Printf("Created user with ID: %d\n", id)
```

#### Read

```go
// Get a user by ID
user, err := orm.Get("users", 1)
if err != nil {
    log.Fatalf("Failed to get user: %v", err)
}
fmt.Printf("User: %v\n", user)

// Query all users
users, err := orm.Query("SELECT * FROM users")
if err != nil {
    log.Fatalf("Failed to query users: %v", err)
}
for _, user := range users {
    fmt.Printf("User: %v\n", user)
}
```

#### Update

```go
// Update a user
err := orm.Update("users", 1, map[string]interface{}{
    "username": "johndoe",
    "email":    "johndoe@example.com",
})
if err != nil {
    log.Fatalf("Failed to update user: %v", err)
}
```

#### Delete

```go
// Delete a user
err := orm.Delete("users", 1)
if err != nil {
    log.Fatalf("Failed to delete user: %v", err)
}
```

### Custom Queries

```go
// Execute a custom query
results, err := orm.Query("SELECT * FROM users WHERE username LIKE ?", "%john%")
if err != nil {
    log.Fatalf("Failed to execute query: %v", err)
}
for _, result := range results {
    fmt.Printf("Result: %v\n", result)
}
```

## Creating Serializers

Serializers transform data between your models and JSON. They also handle validation.

### Basic Serializer

```go
// Create a serializer for the user model
userSerializer := serializer.New(userModel)
```

### Field Customization

```go
// Make a field write-only (won't be included in responses)
userSerializer.AddField("password", reflect.TypeOf(""), serializer.WithWriteOnly())

// Make a field read-only (won't be updated from requests)
userSerializer.AddField("created_at", reflect.TypeOf(time.Time{}), serializer.WithReadOnly())

// Set a default value for a field
userSerializer.AddField("is_active", reflect.TypeOf(true), serializer.WithDefault(true))

// Map a field to a different source field
userSerializer.AddField("full_name", reflect.TypeOf(""), serializer.WithSourceField("name"))
```

### Validation

```go
// Add email validation
userSerializer.AddField("email", reflect.TypeOf(""), 
    serializer.WithValidator(serializer.EmailValidator()))

// Add minimum length validation
userSerializer.AddField("password", reflect.TypeOf(""), 
    serializer.WithValidator(serializer.MinLengthValidator(8)))

// Add maximum length validation
userSerializer.AddField("username", reflect.TypeOf(""), 
    serializer.WithValidator(serializer.MaxLengthValidator(50)))

// Add range validation for numbers
userSerializer.AddField("age", reflect.TypeOf(0), 
    serializer.WithValidator(serializer.RangeValidator(18, 100)))

// Add custom validation
userSerializer.AddField("status", reflect.TypeOf(""), 
    serializer.WithValidator(func(value interface{}) error {
        status, ok := value.(string)
        if !ok {
            return fmt.Errorf("status must be a string")
        }
        validStatuses := []string{"active", "inactive", "pending"}
        for _, validStatus := range validStatuses {
            if status == validStatus {
                return nil
            }
        }
        return fmt.Errorf("invalid status: must be one of %v", validStatuses)
    }))
```

## Creating Views and Routes

### Controllers

Controllers handle HTTP requests and responses. They provide methods for listing, retrieving, creating, updating, and deleting resources.

```go
// Create a controller for the user model
userController := api.NewController(orm, userModel, "/api/users")

// Set a custom serializer for the controller
userController.SetSerializer(userSerializer)
```

### Routing

The router handles HTTP requests and routes them to the appropriate controllers.

```go
// Create a new router
r := router.New()

// Register routes directly
r.GET("/api/users", userController.List)
r.GET("/api/users/:id", userController.Get)
r.POST("/api/users", userController.Create)
r.PUT("/api/users/:id", userController.Update)
r.DELETE("/api/users/:id", userController.Delete)

// Or use route groups
apiGroup := r.Group("/api")
apiGroup.GET("/users", userController.List)
apiGroup.GET("/users/:id", userController.Get)
apiGroup.POST("/users", userController.Create)
apiGroup.PUT("/users/:id", userController.Update)
apiGroup.DELETE("/users/:id", userController.Delete)
```

### Middleware

Middleware adds functionality to your API. It can be used for logging, authentication, CORS, etc.

```go
// Add middleware to the router
r.Use(middleware.Logger)
r.Use(middleware.Recovery)
r.Use(middleware.CORS)

// Add middleware to specific routes
apiGroup.POST("/users", userController.Create, middleware.Auth)
apiGroup.PUT("/users/:id", userController.Update, middleware.Auth)
apiGroup.DELETE("/users/:id", userController.Delete, middleware.Auth)
```

## Configuration

FrameGo provides a configuration system that allows you to manage your application settings.

### Loading Configuration

```go
// Load configuration from file
cfg, err := config.LoadFromFile("config.json")
if err != nil {
    log.Printf("Failed to load config from file: %v", err)
    log.Println("Using default configuration")
    cfg = config.DefaultConfig()
}

// Load configuration from environment variables
cfg := config.LoadFromEnv()
```

### Database Configuration

Example configuration file (config.json):

```json
{
  "database": {
    "driver": "sqlite3",
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "",
    "database": "test.db"
  },
  "server": {
    "host": "localhost",
    "port": 8080
  },
  "debug": true,
  "secret_key": "your-secret-key-here",
  "graphql": {
    "enabled": true,
    "path": "/graphql"
  }
}
```

Using the configuration:

```go
// Create ORM using configuration
orm, err := orm.New(cfg.ToORMConfig())
if err != nil {
    log.Fatalf("Failed to create ORM: %v", err)
}

// Start server using configuration
serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
log.Fatal(http.ListenAndServe(serverAddr, r))
```

## Using GraphQL

FrameGo provides GraphQL support that allows you to expose your models via GraphQL in addition to REST.

### Setting Up GraphQL

```go
// Create a new GraphQL handler
graphqlHandler := graphql.New(orm)

// Register models with GraphQL
if err := graphqlHandler.RegisterModel(userModel); err != nil {
    log.Printf("Failed to register user model with GraphQL: %v", err)
}
if err := graphqlHandler.RegisterModel(postModel); err != nil {
    log.Printf("Failed to register post model with GraphQL: %v", err)
}

// Register GraphQL handler
http.Handle("/graphql", graphqlHandler)
```

### Queries and Mutations

Once you've registered your models with GraphQL, you can query and mutate them using GraphQL syntax:

#### Queries

```graphql
# Get all users
{
  users {
    id
    username
    email
  }
}

# Get a specific user
{
  user(id: 1) {
    id
    username
    email
    posts {
      id
      title
      content
    }
  }
}
```

#### Mutations

```graphql
# Create a new user
mutation {
  createUser(username: "john", email: "john@example.com", password: "password123") {
    id
    username
    email
  }
}

# Update a user
mutation {
  updateUser(id: 1, username: "johndoe", email: "johndoe@example.com") {
    id
    username
    email
  }
}

# Delete a user
mutation {
  deleteUser(id: 1)
}
```

## Examples

### Blog Application

Here's an example of a simple blog application with users, posts, and comments:

```go
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

func main() {
	// Load configuration
	cfg, err := config.LoadFromFile("config.json")
	if err != nil {
		log.Printf("Failed to load config from file: %v", err)
		cfg = config.DefaultConfig()
	}

	// Create ORM
	orm, err := orm.New(cfg.ToORMConfig())
	if err != nil {
		log.Fatalf("Failed to create ORM: %v", err)
	}
	defer orm.Close()

	// Create models
	userModel := models.NewModel("users")
	userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	userModel.AddField("email", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100), models.WithUnique())
	userModel.AddField("password", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	userModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	postModel := models.NewModel("posts")
	postModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	postModel.AddField("title", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	postModel.AddField("content", reflect.TypeOf(""), models.WithNotNull())
	postModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
	postModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	postModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	commentModel := models.NewModel("comments")
	commentModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	commentModel.AddField("content", reflect.TypeOf(""), models.WithNotNull())
	commentModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
	commentModel.AddField("post_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("posts", "id", "CASCADE", "CASCADE"))
	commentModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	commentModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Register models with ORM
	if err := orm.RegisterModel(userModel); err != nil {
		log.Fatalf("Failed to register user model: %v", err)
	}
	if err := orm.RegisterModel(postModel); err != nil {
		log.Fatalf("Failed to register post model: %v", err)
	}
	if err := orm.RegisterModel(commentModel); err != nil {
		log.Fatalf("Failed to register comment model: %v", err)
	}

	// Create tables
	if err := orm.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create serializers
	userSerializer := serializer.New(userModel)
	userSerializer.AddField("password", reflect.TypeOf(""), serializer.WithWriteOnly())
	userSerializer.AddField("email", reflect.TypeOf(""), serializer.WithValidator(serializer.EmailValidator()))

	postSerializer := serializer.New(postModel)
	postSerializer.AddField("title", reflect.TypeOf(""), 
		serializer.WithValidator(serializer.MinLengthValidator(5)))
	postSerializer.AddField("content", reflect.TypeOf(""), 
		serializer.WithValidator(serializer.MinLengthValidator(10)))

	commentSerializer := serializer.New(commentModel)
	commentSerializer.AddField("content", reflect.TypeOf(""), 
		serializer.WithValidator(serializer.MinLengthValidator(1)))

	// Create controllers
	userController := api.NewController(orm, userModel, "/api/users")
	userController.SetSerializer(userSerializer)

	postController := api.NewController(orm, postModel, "/api/posts")
	postController.SetSerializer(postSerializer)

	commentController := api.NewController(orm, commentModel, "/api/comments")
	commentController.SetSerializer(commentSerializer)

	// Create router
	r := router.New()
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// Register routes
	apiGroup := r.Group("/api")

	// User routes
	apiGroup.GET("/users", userController.List)
	apiGroup.GET("/users/:id", userController.Get)
	apiGroup.POST("/users", userController.Create)
	apiGroup.PUT("/users/:id", userController.Update, middleware.Auth)
	apiGroup.DELETE("/users/:id", userController.Delete, middleware.Auth)

	// Post routes
	apiGroup.GET("/posts", postController.List)
	apiGroup.GET("/posts/:id", postController.Get)
	apiGroup.POST("/posts", postController.Create, middleware.Auth)
	apiGroup.PUT("/posts/:id", postController.Update, middleware.Auth)
	apiGroup.DELETE("/posts/:id", postController.Delete, middleware.Auth)

	// Comment routes
	apiGroup.GET("/comments", commentController.List)
	apiGroup.GET("/comments/:id", commentController.Get)
	apiGroup.POST("/comments", commentController.Create, middleware.Auth)
	apiGroup.PUT("/comments/:id", commentController.Update, middleware.Auth)
	apiGroup.DELETE("/comments/:id", commentController.Delete, middleware.Auth)

	// Create GraphQL handler if enabled
	if cfg.GraphQL.Enabled {
		graphqlHandler := graphql.New(orm)

		// Register models with GraphQL
		if err := graphqlHandler.RegisterModel(userModel); err != nil {
			log.Printf("Failed to register user model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(postModel); err != nil {
			log.Printf("Failed to register post model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(commentModel); err != nil {
			log.Printf("Failed to register comment model with GraphQL: %v", err)
		}

		// Register GraphQL handler
		http.Handle(cfg.GraphQL.Path, graphqlHandler)
		log.Printf("GraphQL endpoint available at http://%s:%d%s", cfg.Server.Host, cfg.Server.Port, cfg.GraphQL.Path)
	}

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started at http://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
```

### E-commerce Application

Here's an example of a simple e-commerce application with products, categories, and orders:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"framego/pkg/api"
	"framego/pkg/config"
	"framego/pkg/graphql"
	"framego/pkg/middleware"
	"framego/pkg/models"
	"framego/pkg/orm"
	"framego/pkg/router"
	"framego/pkg/serializer"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromFile("config.json")
	if err != nil {
		log.Printf("Failed to load config from file: %v", err)
		cfg = config.DefaultConfig()
	}

	// Create ORM
	orm, err := orm.New(cfg.ToORMConfig())
	if err != nil {
		log.Fatalf("Failed to create ORM: %v", err)
	}
	defer orm.Close()

	// Create models
	userModel := models.NewModel("users")
	userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	userModel.AddField("email", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100), models.WithUnique())
	userModel.AddField("password", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	userModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	productModel := models.NewModel("products")
	productModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	productModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	productModel.AddField("description", reflect.TypeOf(""), models.WithMaxLength(500))
	productModel.AddField("price", reflect.TypeOf(0.0), models.WithNotNull())
	productModel.AddField("stock", reflect.TypeOf(0), models.WithNotNull(), models.WithDefault(0))
	productModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	productModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	categoryModel := models.NewModel("categories")
	categoryModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	categoryModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	categoryModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	categoryModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	productCategoryModel := models.NewModel("product_categories")
	productCategoryModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	productCategoryModel.AddField("product_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("products", "id", "CASCADE", "CASCADE"))
	productCategoryModel.AddField("category_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("categories", "id", "CASCADE", "CASCADE"))
	productCategoryModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	productCategoryModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	orderModel := models.NewModel("orders")
	orderModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	orderModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
	orderModel.AddField("total_price", reflect.TypeOf(0.0), models.WithNotNull())
	orderModel.AddField("status", reflect.TypeOf(""), models.WithNotNull(), models.WithDefault("pending"))
	orderModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	orderModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	orderItemModel := models.NewModel("order_items")
	orderItemModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	orderItemModel.AddField("order_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("orders", "id", "CASCADE", "CASCADE"))
	orderItemModel.AddField("product_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("products", "id", "RESTRICT", "CASCADE"))
	orderItemModel.AddField("quantity", reflect.TypeOf(0), models.WithNotNull())
	orderItemModel.AddField("price", reflect.TypeOf(0.0), models.WithNotNull())
	orderItemModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	orderItemModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Register models with ORM
	if err := orm.RegisterModel(userModel); err != nil {
		log.Fatalf("Failed to register user model: %v", err)
	}
	if err := orm.RegisterModel(productModel); err != nil {
		log.Fatalf("Failed to register product model: %v", err)
	}
	if err := orm.RegisterModel(categoryModel); err != nil {
		log.Fatalf("Failed to register category model: %v", err)
	}
	if err := orm.RegisterModel(productCategoryModel); err != nil {
		log.Fatalf("Failed to register product_category model: %v", err)
	}
	if err := orm.RegisterModel(orderModel); err != nil {
		log.Fatalf("Failed to register order model: %v", err)
	}
	if err := orm.RegisterModel(orderItemModel); err != nil {
		log.Fatalf("Failed to register order_item model: %v", err)
	}

	// Create tables
	if err := orm.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create serializers
	userSerializer := serializer.New(userModel)
	userSerializer.AddField("password", reflect.TypeOf(""), serializer.WithWriteOnly())
	userSerializer.AddField("email", reflect.TypeOf(""), serializer.WithValidator(serializer.EmailValidator()))

	productSerializer := serializer.New(productModel)
	productSerializer.AddField("price", reflect.TypeOf(0.0), 
		serializer.WithValidator(serializer.RangeValidator(0.01, 1000000.0)))
	productSerializer.AddField("stock", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(0, 1000000)))

	categorySerializer := serializer.New(categoryModel)

	productCategorySerializer := serializer.New(productCategoryModel)

	orderSerializer := serializer.New(orderModel)
	orderSerializer.AddField("status", reflect.TypeOf(""), 
		serializer.WithValidator(func(value interface{}) error {
			status, ok := value.(string)
			if !ok {
				return fmt.Errorf("status must be a string")
			}
			validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
			for _, validStatus := range validStatuses {
				if status == validStatus {
					return nil
				}
			}
			return fmt.Errorf("invalid status: must be one of %v", validStatuses)
		}))

	orderItemSerializer := serializer.New(orderItemModel)
	orderItemSerializer.AddField("quantity", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(1, 100)))

	// Create controllers
	userController := api.NewController(orm, userModel, "/api/users")
	userController.SetSerializer(userSerializer)

	productController := api.NewController(orm, productModel, "/api/products")
	productController.SetSerializer(productSerializer)

	categoryController := api.NewController(orm, categoryModel, "/api/categories")
	categoryController.SetSerializer(categorySerializer)

	productCategoryController := api.NewController(orm, productCategoryModel, "/api/product-categories")
	productCategoryController.SetSerializer(productCategorySerializer)

	orderController := api.NewController(orm, orderModel, "/api/orders")
	orderController.SetSerializer(orderSerializer)

	orderItemController := api.NewController(orm, orderItemModel, "/api/order-items")
	orderItemController.SetSerializer(orderItemSerializer)

	// Create router
	r := router.New()
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS)

	// Register routes
	apiGroup := r.Group("/api")

	// User routes
	apiGroup.GET("/users", userController.List)
	apiGroup.GET("/users/:id", userController.Get)
	apiGroup.POST("/users", userController.Create)
	apiGroup.PUT("/users/:id", userController.Update, middleware.Auth)
	apiGroup.DELETE("/users/:id", userController.Delete, middleware.Auth)

	// Product routes
	apiGroup.GET("/products", productController.List)
	apiGroup.GET("/products/:id", productController.Get)
	apiGroup.POST("/products", productController.Create, middleware.Auth)
	apiGroup.PUT("/products/:id", productController.Update, middleware.Auth)
	apiGroup.DELETE("/products/:id", productController.Delete, middleware.Auth)

	// Category routes
	apiGroup.GET("/categories", categoryController.List)
	apiGroup.GET("/categories/:id", categoryController.Get)
	apiGroup.POST("/categories", categoryController.Create, middleware.Auth)
	apiGroup.PUT("/categories/:id", categoryController.Update, middleware.Auth)
	apiGroup.DELETE("/categories/:id", categoryController.Delete, middleware.Auth)

	// ProductCategory routes
	apiGroup.GET("/product-categories", productCategoryController.List)
	apiGroup.GET("/product-categories/:id", productCategoryController.Get)
	apiGroup.POST("/product-categories", productCategoryController.Create, middleware.Auth)
	apiGroup.PUT("/product-categories/:id", productCategoryController.Update, middleware.Auth)
	apiGroup.DELETE("/product-categories/:id", productCategoryController.Delete, middleware.Auth)

	// Order routes
	apiGroup.GET("/orders", orderController.List, middleware.Auth)
	apiGroup.GET("/orders/:id", orderController.Get, middleware.Auth)
	apiGroup.POST("/orders", orderController.Create, middleware.Auth)
	apiGroup.PUT("/orders/:id", orderController.Update, middleware.Auth)
	apiGroup.DELETE("/orders/:id", orderController.Delete, middleware.Auth)

	// OrderItem routes
	apiGroup.GET("/order-items", orderItemController.List, middleware.Auth)
	apiGroup.GET("/order-items/:id", orderItemController.Get, middleware.Auth)
	apiGroup.POST("/order-items", orderItemController.Create, middleware.Auth)
	apiGroup.PUT("/order-items/:id", orderItemController.Update, middleware.Auth)
	apiGroup.DELETE("/order-items/:id", orderItemController.Delete, middleware.Auth)

	// Create GraphQL handler if enabled
	if cfg.GraphQL.Enabled {
		graphqlHandler := graphql.New(orm)

		// Register models with GraphQL
		if err := graphqlHandler.RegisterModel(userModel); err != nil {
			log.Printf("Failed to register user model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(productModel); err != nil {
			log.Printf("Failed to register product model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(categoryModel); err != nil {
			log.Printf("Failed to register category model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(productCategoryModel); err != nil {
			log.Printf("Failed to register product_category model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(orderModel); err != nil {
			log.Printf("Failed to register order model with GraphQL: %v", err)
		}
		if err := graphqlHandler.RegisterModel(orderItemModel); err != nil {
			log.Printf("Failed to register order_item model with GraphQL: %v", err)
		}

		// Register GraphQL handler
		http.Handle(cfg.GraphQL.Path, graphqlHandler)
		log.Printf("GraphQL endpoint available at http://%s:%d%s", cfg.Server.Host, cfg.Server.Port, cfg.GraphQL.Path)
	}

	// Start server
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Server started at http://%s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
```

## License

MIT
