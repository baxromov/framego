package orders

import (
	"reflect"
	"time"

	"framego/pkg/models"
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

// CreateOrderModel creates and returns an order model
func CreateOrderModel() *models.Model {
	orderModel := models.NewModel("orders")
	orderModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	orderModel.AddField("user_id", reflect.TypeOf(0), models.WithNotNull(), 
		models.WithForeignKey("users", "id", "CASCADE", "CASCADE"))
	orderModel.AddField("total_price", reflect.TypeOf(0.0), models.WithNotNull())
	orderModel.AddField("status", reflect.TypeOf(""), models.WithNotNull(), models.WithDefault("pending"))
	orderModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	orderModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	
	return orderModel
}

// CreateOrderItemModel creates and returns an order item model
func CreateOrderItemModel() *models.Model {
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
	
	return orderItemModel
}