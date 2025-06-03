package orders

import (
	"fmt"

	"github.com/baxromov/framego/pkg/api"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
)

// SetupOrderAPI sets up the order API routes
func SetupOrderAPI(orm *orm.ORM, r *router.Router) {
	// Create order model
	orderModel := CreateOrderModel()

	// Create order item model
	orderItemModel := CreateOrderItemModel()

	// Register models with ORM
	if err := orm.RegisterModel(orderModel); err != nil {
		panic(err)
	}
	if err := orm.RegisterModel(orderItemModel); err != nil {
		panic(err)
	}

	// Create order controller
	orderController := api.NewController(orm, orderModel, "/api/orders")
	orderSerializer := CreateOrderSerializer(orderModel)
	orderController.SetSerializer(orderSerializer)

	// Create order item controller
	orderItemController := api.NewController(orm, orderItemModel, "/api/order-items")
	orderItemSerializer := CreateOrderItemSerializer(orderItemModel)
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
}
