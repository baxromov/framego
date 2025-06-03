package users

import (
	"reflect"

	"framego/pkg/models"
	"framego/pkg/serializer"
)

// CreateUserSerializer creates and returns a user serializer
func CreateUserSerializer(userModel *models.Model) *serializer.Serializer {
	userSerializer := serializer.New(userModel)
	
	// Make password write-only
	userSerializer.AddField("password", reflect.TypeOf(""), serializer.WithWriteOnly())
	
	// Add email validator
	userSerializer.AddField("email", reflect.TypeOf(""), serializer.WithValidator(serializer.EmailValidator()))
	
	return userSerializer
}