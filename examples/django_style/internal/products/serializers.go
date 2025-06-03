package products

import (
	"reflect"

	"framego/pkg/models"
	"framego/pkg/serializer"
)

// CreateProductSerializer creates and returns a product serializer
func CreateProductSerializer(productModel *models.Model) *serializer.Serializer {
	productSerializer := serializer.New(productModel)
	
	// Add price validator (must be positive)
	productSerializer.AddField("price", reflect.TypeOf(0.0), 
		serializer.WithValidator(serializer.RangeValidator(0.01, 1000000.0)))
	
	// Add stock validator (must be non-negative)
	productSerializer.AddField("stock", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(0, 1000000)))
	
	return productSerializer
}