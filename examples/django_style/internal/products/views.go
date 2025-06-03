package products

import (
	"fmt"

	"github.com/baxromov/framego/pkg/api"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
)

// SetupProductAPI sets up the product API routes
func SetupProductAPI(orm *orm.ORM, r *router.Router) {
	// Create product model
	productModel := CreateProductModel()

	// Register the product model with the ORM
	if err := orm.RegisterModel(productModel); err != nil {
		panic(err)
	}

	// Create a new API controller for the product model
	productController := api.NewController(orm, productModel, "/api/products")

	// Create a custom serializer for the product model
	productSerializer := CreateProductSerializer(productModel)

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
}
