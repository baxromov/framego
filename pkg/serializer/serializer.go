package serializer

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/baxromov/framego/pkg/models"
)

// Field represents a serializer field
type Field struct {
	Name          string
	Type          reflect.Type
	Required      bool
	ReadOnly      bool
	WriteOnly     bool
	Default       interface{}
	Validators    []Validator
	SourceField   string
	ErrorMessages map[string]string
}

// Validator defines a function that validates a field value
type Validator func(value interface{}) error

// Serializer represents a serializer for a model
type Serializer struct {
	Model  models.ModelInterface
	Fields map[string]Field
}

// New creates a new serializer for the given model
func New(model models.ModelInterface) *Serializer {
	s := &Serializer{
		Model:  model,
		Fields: make(map[string]Field),
	}

	// Auto-generate fields from model
	for name, modelField := range model.GetFields() {
		field := Field{
			Name:        name,
			Type:        modelField.Type,
			Required:    modelField.NotNull && modelField.Default == nil,
			SourceField: name,
		}

		// Add validators based on model field constraints
		if modelField.MaxLength > 0 && modelField.Type.Kind() == reflect.String {
			field.Validators = append(field.Validators, MaxLengthValidator(modelField.MaxLength))
		}

		s.Fields[name] = field
	}

	return s
}

// AddField adds a field to the serializer
func (s *Serializer) AddField(name string, fieldType reflect.Type, options ...func(*Field)) {
	field := Field{
		Name:        name,
		Type:        fieldType,
		SourceField: name,
	}

	// Apply options
	for _, option := range options {
		option(&field)
	}

	s.Fields[name] = field
}

// WithRequired sets the field as required
func WithRequired() func(*Field) {
	return func(f *Field) {
		f.Required = true
	}
}

// WithReadOnly sets the field as read-only
func WithReadOnly() func(*Field) {
	return func(f *Field) {
		f.ReadOnly = true
	}
}

// WithWriteOnly sets the field as write-only
func WithWriteOnly() func(*Field) {
	return func(f *Field) {
		f.WriteOnly = true
	}
}

// WithDefault sets the default value for the field
func WithDefault(value interface{}) func(*Field) {
	return func(f *Field) {
		f.Default = value
	}
}

// WithValidator adds a validator to the field
func WithValidator(validator Validator) func(*Field) {
	return func(f *Field) {
		f.Validators = append(f.Validators, validator)
	}
}

// WithSourceField sets the source field for the field
func WithSourceField(sourceField string) func(*Field) {
	return func(f *Field) {
		f.SourceField = sourceField
	}
}

// WithErrorMessages sets custom error messages for the field
func WithErrorMessages(messages map[string]string) func(*Field) {
	return func(f *Field) {
		f.ErrorMessages = messages
	}
}

// Serialize converts a model instance to a map
func (s *Serializer) Serialize(data interface{}) (map[string]interface{}, error) {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return nil, fmt.Errorf("data must be a struct, a map, or a pointer to a struct")
	}

	result := make(map[string]interface{})

	// Handle struct
	if val.Kind() == reflect.Struct {
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

			// Check if field is in serializer fields
			if serializerField, ok := s.Fields[name]; ok {
				if serializerField.WriteOnly {
					continue // Skip write-only fields
				}
				result[name] = val.Field(i).Interface()
			}
		}
	} else if val.Kind() == reflect.Map {
		// Handle map
		for _, key := range val.MapKeys() {
			name := key.String()
			if serializerField, ok := s.Fields[name]; ok {
				if serializerField.WriteOnly {
					continue // Skip write-only fields
				}
				result[name] = val.MapIndex(key).Interface()
			}
		}
	}

	return result, nil
}

// Deserialize converts a map to a model instance
func (s *Serializer) Deserialize(data map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})

	// Apply defaults
	for name, field := range s.Fields {
		if field.ReadOnly {
			continue // Skip read-only fields
		}

		// If field has a default value and is not in data, use default
		if field.Default != nil {
			if _, ok := data[name]; !ok {
				result[field.SourceField] = field.Default
			}
		}
	}

	// Copy data to result, validating as we go
	for name, value := range data {
		field, ok := s.Fields[name]
		if !ok {
			continue // Skip fields not in serializer
		}

		if field.ReadOnly {
			continue // Skip read-only fields
		}

		// Validate field
		if err := s.validateField(name, value); err != nil {
			return nil, err
		}

		result[field.SourceField] = value
	}

	// Check required fields
	for name, field := range s.Fields {
		if field.Required {
			if _, ok := result[field.SourceField]; !ok {
				return nil, fmt.Errorf("field %s is required", name)
			}
		}
	}

	return result, nil
}

// validateField validates a field value
func (s *Serializer) validateField(name string, value interface{}) error {
	field, ok := s.Fields[name]
	if !ok {
		return fmt.Errorf("field %s not found", name)
	}

	// Type validation
	valueType := reflect.TypeOf(value)
	if valueType != field.Type && value != nil {
		// Special case for time.Time
		if field.Type == reflect.TypeOf(time.Time{}) {
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field %s has invalid type: expected time.Time or string, got %v", name, valueType)
			}
			// TODO: Parse string to time.Time
		} else {
			return fmt.Errorf("field %s has invalid type: expected %v, got %v", name, field.Type, valueType)
		}
	}

	// Run validators
	for _, validator := range field.Validators {
		if err := validator(value); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the data against the serializer's constraints
func (s *Serializer) Validate(data map[string]interface{}) error {
	// Check required fields
	for name, field := range s.Fields {
		if field.Required {
			if _, ok := data[name]; !ok {
				return fmt.Errorf("field %s is required", name)
			}
		}
	}

	// Validate fields
	for name, value := range data {
		if err := s.validateField(name, value); err != nil {
			return err
		}
	}

	return nil
}

// Common validators

// MaxLengthValidator creates a validator that checks if a string exceeds a maximum length
func MaxLengthValidator(maxLength int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not a string")
		}
		if len(str) > maxLength {
			return fmt.Errorf("string length exceeds maximum of %d", maxLength)
		}
		return nil
	}
}

// MinLengthValidator creates a validator that checks if a string meets a minimum length
func MinLengthValidator(minLength int) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not a string")
		}
		if len(str) < minLength {
			return fmt.Errorf("string length is less than minimum of %d", minLength)
		}
		return nil
	}
}

// RegexValidator creates a validator that checks if a string matches a regex pattern
func RegexValidator(pattern string) Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not a string")
		}
		// TODO: Implement regex validation
		_ = str
		return nil
	}
}

// EmailValidator creates a validator that checks if a string is a valid email
func EmailValidator() Validator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("value is not a string")
		}
		if !strings.Contains(str, "@") {
			return fmt.Errorf("invalid email address")
		}
		return nil
	}
}

// RangeValidator creates a validator that checks if a number is within a range
func RangeValidator(min, max float64) Validator {
	return func(value interface{}) error {
		var num float64
		switch v := value.(type) {
		case int:
			num = float64(v)
		case int8:
			num = float64(v)
		case int16:
			num = float64(v)
		case int32:
			num = float64(v)
		case int64:
			num = float64(v)
		case uint:
			num = float64(v)
		case uint8:
			num = float64(v)
		case uint16:
			num = float64(v)
		case uint32:
			num = float64(v)
		case uint64:
			num = float64(v)
		case float32:
			num = float64(v)
		case float64:
			num = v
		default:
			return fmt.Errorf("value is not a number")
		}

		if num < min || num > max {
			return fmt.Errorf("number must be between %f and %f", min, max)
		}
		return nil
	}
}
