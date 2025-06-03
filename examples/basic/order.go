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

// Order represents an order model
type Order struct {
	models.Model
	ID         int
	UserID     int
	TotalPrice float64
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// OrderItem represents an order item model
type OrderItem struct {
	models.Model
	ID        int
	OrderID   int
	ProductID int
	Quantity  int
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func setupOrderAPI(orm *orm.ORM, r *router.Router, graphqlHandler *graphql.Handler) {
	// Create order model
	orderModel := models.NewModel("orders")
	orderModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	orderModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
	orderModel.AddField("total_price", reflect.TypeOf(0.0), models.WithNotNull())
	orderModel.AddField("status", reflect.TypeOf(""), models.WithNotNull(), models.WithDefault("pending"))
	orderModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	orderModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	// Create order item model
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
	if err := orm.RegisterModel(orderModel); err != nil {
		log.Fatalf("Failed to register order model: %v", err)
	}
	if err := orm.RegisterModel(orderItemModel); err != nil {
		log.Fatalf("Failed to register order item model: %v", err)
	}

	// Create order controller
	orderController := api.NewController(orm, orderModel, "/api/orders")
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
	orderController.SetSerializer(orderSerializer)

	// Create order item controller
	orderItemController := api.NewController(orm, orderItemModel, "/api/order-items")
	orderItemSerializer := serializer.New(orderItemModel)
	orderItemSerializer.AddField("quantity", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(1, 100)))
	orderItemController.SetSerializer(orderItemSerializer)

	// Register routes
	apiGroup := r.Group("/api")

	// Order routes
	apiGroup.GET("/orders", orderController.List, middleware.Auth)
	apiGroup.GET("/orders/:id", orderController.Get, middleware.Auth)
	apiGroup.POST("/orders", orderController.Create, middleware.Auth)
	apiGroup.PUT("/orders/:id", orderController.Update, middleware.Auth)
	apiGroup.DELETE("/orders/:id", orderController.Delete, middleware.Auth)

	// Order item routes
	apiGroup.GET("/order-items", orderItemController.List, middleware.Auth)
	apiGroup.GET("/order-items/:id", orderItemController.Get, middleware.Auth)
	apiGroup.POST("/order-items", orderItemController.Create, middleware.Auth)
	apiGroup.PUT("/order-items/:id", orderItemController.Update, middleware.Auth)
	apiGroup.DELETE("/order-items/:id", orderItemController.Delete, middleware.Auth)

	fmt.Println("Order API routes registered")

	// Register with GraphQL if handler is provided
	if graphqlHandler != nil {
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
	}
}
