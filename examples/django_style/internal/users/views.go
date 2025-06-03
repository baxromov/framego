package users

import (
	"github.com/baxromov/framego/pkg/api"
	"github.com/baxromov/framego/pkg/middleware"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/baxromov/framego/pkg/router"
)

// SetupUserAPI sets up the user API routes
func SetupUserAPI(orm *orm.ORM, r *router.Router) {
	// Create user model
	userModel := CreateUserModel()

	// Register the user model with the ORM
	if err := orm.RegisterModel(userModel); err != nil {
		panic(err)
	}

	// Create a new API controller for the user model
	userController := api.NewController(orm, userModel, "/api/users")

	// Create a custom serializer for the user model
	userSerializer := CreateUserSerializer(userModel)

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
}
