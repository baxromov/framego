package orders

import (
	"fmt"
	"reflect"

	"framego/pkg/models"
	"framego/pkg/serializer"
)

// CreateOrderSerializer creates and returns an order serializer
func CreateOrderSerializer(orderModel *models.Model) *serializer.Serializer {
	orderSerializer := serializer.New(orderModel)
	
	// Add status validator
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
	
	return orderSerializer
}

// CreateOrderItemSerializer creates and returns an order item serializer
func CreateOrderItemSerializer(orderItemModel *models.Model) *serializer.Serializer {
	orderItemSerializer := serializer.New(orderItemModel)
	
	// Add quantity validator (must be positive)
	orderItemSerializer.AddField("quantity", reflect.TypeOf(0), 
		serializer.WithValidator(serializer.RangeValidator(1, 100)))
	
	return orderItemSerializer
}