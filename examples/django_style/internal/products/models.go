package products

import (
	"reflect"
	"time"

	"github.com/baxromov/framego/pkg/models"
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

// CreateProductModel creates and returns a product model
func CreateProductModel() *models.Model {
	productModel := models.NewModel("products")
	productModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	productModel.AddField("name", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	productModel.AddField("description", reflect.TypeOf(""), models.WithMaxLength(500))
	productModel.AddField("price", reflect.TypeOf(0.0), models.WithNotNull())
	productModel.AddField("stock", reflect.TypeOf(0), models.WithNotNull(), models.WithDefault(0))
	productModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	productModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))

	return productModel
}
