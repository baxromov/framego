package users

import (
	"reflect"
	"time"

	"framego/pkg/models"
)

// User represents a user model
type User struct {
	models.Model
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateUserModel creates and returns a user model
func CreateUserModel() *models.Model {
	userModel := models.NewModel("users")
	userModel.AddField("id", reflect.TypeOf(0), models.WithPrimaryKey(), models.WithAutoIncrement())
	userModel.AddField("username", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(50), models.WithUnique())
	userModel.AddField("email", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100), models.WithUnique())
	userModel.AddField("password", reflect.TypeOf(""), models.WithNotNull(), models.WithMaxLength(100))
	userModel.AddField("created_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	userModel.AddField("updated_at", reflect.TypeOf(time.Time{}), models.WithNotNull(), models.WithDefault(time.Now()))
	
	return userModel
}