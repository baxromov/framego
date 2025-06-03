package models

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Field represents a database column
type Field struct {
	Name         string
	Type         reflect.Type
	PrimaryKey   bool
	AutoIncrement bool
	Unique       bool
	NotNull      bool
	Default      interface{}
	MaxLength    int
	ForeignKey   *ForeignKey
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Model      string
	Field      string
	OnDelete   string
	OnUpdate   string
}

// Model is the base struct for all models
type Model struct {
	TableName string
	Fields    map[string]Field
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ModelInterface defines methods that all models must implement
type ModelInterface interface {
	GetTableName() string
	GetFields() map[string]Field
	Validate() error
}

// NewModel creates a new model with the given table name
func NewModel(tableName string) *Model {
	return &Model{
		TableName: tableName,
		Fields:    make(map[string]Field),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// GetTableName returns the table name for the model
func (m *Model) GetTableName() string {
	return m.TableName
}

// GetFields returns the fields for the model
func (m *Model) GetFields() map[string]Field {
	return m.Fields
}

// Validate validates the model
func (m *Model) Validate() error {
	// Check if the model has at least one field
	if len(m.Fields) == 0 {
		return fmt.Errorf("model %s has no fields", m.TableName)
	}

	// Check if the model has a primary key
	hasPrimaryKey := false
	for _, field := range m.Fields {
		if field.PrimaryKey {
			hasPrimaryKey = true
			break
		}
	}

	if !hasPrimaryKey {
		return fmt.Errorf("model %s has no primary key", m.TableName)
	}

	return nil
}

// AddField adds a field to the model
func (m *Model) AddField(name string, fieldType reflect.Type, options ...func(*Field)) {
	field := Field{
		Name: name,
		Type: fieldType,
	}

	// Apply options
	for _, option := range options {
		option(&field)
	}

	m.Fields[name] = field
}

// WithPrimaryKey sets the field as a primary key
func WithPrimaryKey() func(*Field) {
	return func(f *Field) {
		f.PrimaryKey = true
	}
}

// WithAutoIncrement sets the field as auto-incrementing
func WithAutoIncrement() func(*Field) {
	return func(f *Field) {
		f.AutoIncrement = true
	}
}

// WithUnique sets the field as unique
func WithUnique() func(*Field) {
	return func(f *Field) {
		f.Unique = true
	}
}

// WithNotNull sets the field as not null
func WithNotNull() func(*Field) {
	return func(f *Field) {
		f.NotNull = true
	}
}

// WithDefault sets the default value for the field
func WithDefault(value interface{}) func(*Field) {
	return func(f *Field) {
		f.Default = value
	}
}

// WithMaxLength sets the maximum length for the field
func WithMaxLength(length int) func(*Field) {
	return func(f *Field) {
		f.MaxLength = length
	}
}

// WithForeignKey sets a foreign key relationship for the field
func WithForeignKey(model, field string, onDelete, onUpdate string) func(*Field) {
	return func(f *Field) {
		f.ForeignKey = &ForeignKey{
			Model:    model,
			Field:    field,
			OnDelete: onDelete,
			OnUpdate: onUpdate,
		}
	}
}

// String returns a string representation of the model
func (m *Model) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Model: %s\n", m.TableName))
	sb.WriteString("Fields:\n")
	
	for name, field := range m.Fields {
		sb.WriteString(fmt.Sprintf("  %s: %v", name, field.Type))
		if field.PrimaryKey {
			sb.WriteString(" (PK)")
		}
		if field.AutoIncrement {
			sb.WriteString(" (AI)")
		}
		if field.Unique {
			sb.WriteString(" (U)")
		}
		if field.NotNull {
			sb.WriteString(" (NN)")
		}
		if field.Default != nil {
			sb.WriteString(fmt.Sprintf(" (Default: %v)", field.Default))
		}
		if field.MaxLength > 0 {
			sb.WriteString(fmt.Sprintf(" (MaxLen: %d)", field.MaxLength))
		}
		if field.ForeignKey != nil {
			sb.WriteString(fmt.Sprintf(" (FK: %s.%s)", field.ForeignKey.Model, field.ForeignKey.Field))
		}
		sb.WriteString("\n")
	}
	
	return sb.String()
}