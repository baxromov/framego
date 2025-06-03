package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"framego/pkg/api"
	"framego/pkg/graphql"
	"framego/pkg/middleware"
	"framego/pkg/models"
	"framego/pkg/orm"
	"framego/pkg/router"
	"framego/pkg/serializer"
)

// Product represents a product model
type Product struct {
	models.Model
	ID          int
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func setupProductAPI(orm *orm.ORM, r *router.Router, graphqlHandler *graphql.Handler) {
	// Create a new product model
	productModel := models.NewModel("products")
	productModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	productModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	productModel.AddField("description", reflect.TypeOf(""), models.WithMaxLength(500))
	productModel.AddField("price", reflect.TypeOf(0.0), models.WithNotNull())
	productModel.AddField("stock", reflect.TypeOf(0), models.WithNotNull(), models.WithDefault(0))
	productModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	productModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Register the product model with the ORM
	if err := orm.RegisterModel(productModel); err != nil {
		log.Fatalf("Failed to register product model: %v", err)
	}

	// Create a new API controller for the product model
	productController := api.NewController(orm, productModel, "/api/products")

	// Create a custom serializer for the product model
	productSerializer := serializer.New(productModel)

	// Add price validator (must be positive)
	productSerializer.AddField("price", reflect.TypeOf(0.0), 
		serializer.WithValidator(serializer.RangeValidator(0.01, 1000000.0)))

	// Add stock validator (must be non-negative)
	productSerializer.AddField("stock", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(0, 1000000)))

	// Set the custom serializer for the controller
	productController.SetSerializer(productSerializer)

	// Register routes
	apiGroup := r.Group("/api")

	// Public routes
	apiGroup.GET("/products", productController.List)
	apiGroup.GET("/products/:id", productController.Get)

	// Protected routes
	apiGroup.POST("/products", productController.Create, middleware.Auth)
	apiGroup.PUT("/products/:id", productController.Update, middleware.Auth)
	apiGroup.DELETE("/products/:id", productController.Delete, middleware.Auth)

	fmt.Println("Product API routes registered")

	// Register with GraphQL if handler is provided
	if graphqlHandler != nil {
		if err := graphqlHandler.RegisterModel(productModel); err != nil {
			log.Printf("Failed to register product model with GraphQL: %v", err)
		} else {
			fmt.Println("Product model registered with GraphQL")
		}
	}
}
