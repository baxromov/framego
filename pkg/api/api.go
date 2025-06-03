package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"framego/pkg/models"
	"framego/pkg/orm"
)

// Controller represents a REST API controller
type Controller struct {
	ORM        *orm.ORM
	Model      models.ModelInterface
	Serializer Serializer
	BasePath   string
}

// Serializer defines methods for serializing and deserializing data
type Serializer interface {
	Serialize(data interface{}) (map[string]interface{}, error)
	Deserialize(data map[string]interface{}) (interface{}, error)
	Validate(data map[string]interface{}) error
}

// DefaultSerializer is a basic implementation of the Serializer interface
type DefaultSerializer struct {
	Model models.ModelInterface
}

// Serialize converts a model instance to a map
func (s *DefaultSerializer) Serialize(data interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data must be a struct or a pointer to a struct")
	}

	result := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields
		}

		// Use JSON tag if available, otherwise use field name
		name := field.Name
		tag := field.Tag.Get("json")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "-" {
				name = parts[0]
			} else {
				continue // Skip fields with json:"-"
			}
		}

		result[name] = val.Field(i).Interface()
	}

	return result, nil
}

// Deserialize converts a map to a model instance
func (s *DefaultSerializer) Deserialize(data map[string]interface{}) (interface{}, error) {
	// This is a simplified implementation
	// In a real-world scenario, you would create a new instance of the model
	// and populate its fields from the data map
	return data, nil
}

// Validate validates the data against the model's constraints
func (s *DefaultSerializer) Validate(data map[string]interface{}) error {
	fields := s.Model.GetFields()

	for name, field := range fields {
		value, exists := data[name]
		if !exists && field.NotNull && field.Default == nil {
			return fmt.Errorf("field %s is required", name)
		}

		if exists {
			// Type validation
			valueType := reflect.TypeOf(value)
			if valueType != field.Type && value != nil {
				return fmt.Errorf("field %s has invalid type: expected %v, got %v", name, field.Type, valueType)
			}

			// String length validation
			if field.Type.Kind() == reflect.String && field.MaxLength > 0 {
				strValue, ok := value.(string)
				if ok && len(strValue) > field.MaxLength {
					return fmt.Errorf("field %s exceeds maximum length of %d", name, field.MaxLength)
				}
			}
		}
	}

	return nil
}

// NewController creates a new controller for the given model
func NewController(orm *orm.ORM, model models.ModelInterface, basePath string) *Controller {
	return &Controller{
		ORM:        orm,
		Model:      model,
		Serializer: &DefaultSerializer{Model: model},
		BasePath:   basePath,
	}
}

// SetSerializer sets a custom serializer for the controller
func (c *Controller) SetSerializer(serializer Serializer) {
	c.Serializer = serializer
}

// RegisterRoutes registers the controller's routes with the given router
func (c *Controller) RegisterRoutes(router http.Handler) {
	// This is a placeholder
	// In a real implementation, you would register routes with the router
	// For example:
	// router.HandleFunc(c.BasePath, c.List).Methods("GET")
	// router.HandleFunc(c.BasePath+"/{id}", c.Get).Methods("GET")
	// router.HandleFunc(c.BasePath, c.Create).Methods("POST")
	// router.HandleFunc(c.BasePath+"/{id}", c.Update).Methods("PUT")
	// router.HandleFunc(c.BasePath+"/{id}", c.Delete).Methods("DELETE")
}

// List handles GET requests to list all records
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	// Query the database for all records
	results, err := c.ORM.Query(fmt.Sprintf("SELECT * FROM %s", c.Model.GetTableName()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize the results
	response := make([]map[string]interface{}, len(results))
	for i, result := range results {
		serialized, err := c.Serializer.Serialize(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response[i] = serialized
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get handles GET requests to retrieve a single record
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL
	id := extractIDFromURL(r.URL.Path)
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Query the database for the record
	result, err := c.ORM.Get(c.Model.GetTableName(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Serialize the result
	response, err := c.Serializer.Serialize(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create handles POST requests to create a new record
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the data
	if err := c.Serializer.Validate(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create the record
	id, err := c.ORM.Create(c.Model.GetTableName(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

// Update handles PUT requests to update an existing record
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL
	id := extractIDFromURL(r.URL.Path)
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse the request body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the data
	if err := c.Serializer.Validate(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the record
	if err := c.ORM.Update(c.Model.GetTableName(), id, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE requests to delete a record
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL
	id := extractIDFromURL(r.URL.Path)
	if id == "" {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Delete the record
	if err := c.ORM.Delete(c.Model.GetTableName(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.WriteHeader(http.StatusNoContent)
}

// extractIDFromURL extracts the ID from the URL path
func extractIDFromURL(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}