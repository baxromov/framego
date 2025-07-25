package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/baxromov/framego/pkg/models"
	"github.com/baxromov/framego/pkg/orm"
	"github.com/graphql-go/graphql"
)

// Handler represents a GraphQL handler
type Handler struct {
	ORM    *orm.ORM
	Models map[string]models.ModelInterface
	Schema *Schema
	// graphql-go schema
	GQLSchema graphql.Schema
}

// Schema represents a GraphQL schema
type Schema struct {
	Types        map[string]*Type
	QueryType    *Type
	MutationType *Type
}

// Type represents a GraphQL type
type Type struct {
	Name        string
	Description string
	Fields      map[string]*Field
	Model       models.ModelInterface
}

// Field represents a GraphQL field
type Field struct {
	Name        string
	Description string
	Type        string
	Args        map[string]*Argument
	Resolve     ResolveFunc
}

// Argument represents a GraphQL argument
type Argument struct {
	Name         string
	Description  string
	Type         string
	DefaultValue interface{}
}

// ResolveFunc is a function that resolves a GraphQL field
type ResolveFunc func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error)

// New creates a new GraphQL handler
func New(orm *orm.ORM) *Handler {
	// Create a new handler
	handler := &Handler{
		ORM:    orm,
		Models: make(map[string]models.ModelInterface),
		Schema: &Schema{
			Types:        make(map[string]*Type),
			QueryType:    &Type{Name: "Query", Fields: make(map[string]*Field)},
			MutationType: &Type{Name: "Mutation", Fields: make(map[string]*Field)},
		},
	}

	// Initialize graphql-go schema with empty query and mutation types
	queryFields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	mutationFields := graphql.Fields{}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: mutationFields,
	})

	schemaConfig := graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		// Log the error but continue
		fmt.Printf("Error creating GraphQL schema: %v\n", err)
	}

	handler.GQLSchema = schema

	return handler
}

// RegisterModel registers a model with the GraphQL handler
func (h *Handler) RegisterModel(model models.ModelInterface) error {
	tableName := model.GetTableName()
	h.Models[tableName] = model

	// Create GraphQL type for model
	typeName := strings.Title(tableName)
	if strings.HasSuffix(typeName, "s") {
		typeName = typeName[:len(typeName)-1]
	}

	modelType := &Type{
		Name:        typeName,
		Description: fmt.Sprintf("Represents a %s", typeName),
		Fields:      make(map[string]*Field),
		Model:       model,
	}

	// Add fields to type
	for name, field := range model.GetFields() {
		modelType.Fields[name] = &Field{
			Name:        name,
			Description: fmt.Sprintf("%s field", name),
			Type:        getGraphQLType(field.Type),
			Args:        make(map[string]*Argument),
			Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
				// If source is a map, return the field value
				if sourceMap, ok := source.(map[string]interface{}); ok {
					return sourceMap[name], nil
				}
				return nil, fmt.Errorf("invalid source type")
			},
		}
	}

	h.Schema.Types[typeName] = modelType

	// Add query fields for model
	h.Schema.QueryType.Fields[tableName] = &Field{
		Name:        tableName,
		Description: fmt.Sprintf("Get all %s", tableName),
		Type:        fmt.Sprintf("[%s]", typeName),
		Args:        make(map[string]*Argument),
		Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
			// Query all records
			results, err := h.ORM.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
			if err != nil {
				return nil, err
			}
			return results, nil
		},
	}

	h.Schema.QueryType.Fields[strings.TrimSuffix(tableName, "s")] = &Field{
		Name:        strings.TrimSuffix(tableName, "s"),
		Description: fmt.Sprintf("Get a %s by ID", strings.TrimSuffix(tableName, "s")),
		Type:        typeName,
		Args: map[string]*Argument{
			"id": {
				Name:        "id",
				Description: "ID of the record",
				Type:        "ID",
			},
		},
		Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
			// Get record by ID
			id, ok := args["id"]
			if !ok {
				return nil, fmt.Errorf("id is required")
			}
			result, err := h.ORM.Get(tableName, id)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
	}

	// Add mutation fields for model
	h.Schema.MutationType.Fields[fmt.Sprintf("create%s", typeName)] = &Field{
		Name:        fmt.Sprintf("create%s", typeName),
		Description: fmt.Sprintf("Create a new %s", typeName),
		Type:        typeName,
		Args:        createInputArgs(model),
		Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
			// Create record
			id, err := h.ORM.Create(tableName, args)
			if err != nil {
				return nil, err
			}
			// Get created record
			result, err := h.ORM.Get(tableName, id)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
	}

	h.Schema.MutationType.Fields[fmt.Sprintf("update%s", typeName)] = &Field{
		Name:        fmt.Sprintf("update%s", typeName),
		Description: fmt.Sprintf("Update an existing %s", typeName),
		Type:        typeName,
		Args:        createInputArgs(model),
		Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
			// Get ID from args
			id, ok := args["id"]
			if !ok {
				return nil, fmt.Errorf("id is required")
			}
			delete(args, "id")

			// Update record
			if err := h.ORM.Update(tableName, id, args); err != nil {
				return nil, err
			}
			// Get updated record
			result, err := h.ORM.Get(tableName, id)
			if err != nil {
				return nil, err
			}
			return result, nil
		},
	}

	h.Schema.MutationType.Fields[fmt.Sprintf("delete%s", typeName)] = &Field{
		Name:        fmt.Sprintf("delete%s", typeName),
		Description: fmt.Sprintf("Delete a %s", typeName),
		Type:        "Boolean",
		Args: map[string]*Argument{
			"id": {
				Name:        "id",
				Description: "ID of the record to delete",
				Type:        "ID",
			},
		},
		Resolve: func(ctx context.Context, source interface{}, args map[string]interface{}) (interface{}, error) {
			// Get ID from args
			id, ok := args["id"]
			if !ok {
				return nil, fmt.Errorf("id is required")
			}

			// Delete record
			if err := h.ORM.Delete(tableName, id); err != nil {
				return nil, err
			}
			return true, nil
		},
	}

	return nil
}

// createInputArgs creates GraphQL arguments for a model
func createInputArgs(model models.ModelInterface) map[string]*Argument {
	args := make(map[string]*Argument)
	for name, field := range model.GetFields() {
		args[name] = &Argument{
			Name:        name,
			Description: fmt.Sprintf("%s field", name),
			Type:        getGraphQLType(field.Type),
		}
	}
	return args
}

// getGraphQLType converts a Go type to a GraphQL type
func getGraphQLType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "Boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "Int"
	case reflect.Int64, reflect.Uint64:
		return "Int"
	case reflect.Float32, reflect.Float64:
		return "Float"
	case reflect.String:
		return "String"
	default:
		if t == reflect.TypeOf(time.Time{}) {
			return "String"
		}
		return "String"
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Accept both POST and GET requests
	var query string
	var variables map[string]interface{}

	switch r.Method {
	case http.MethodPost:
		// Parse POST request body
		var request struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		query = request.Query
		variables = request.Variables
	case http.MethodGet:
		// Parse GET request query parameters
		query = r.URL.Query().Get("query")
		variablesStr := r.URL.Query().Get("variables")
		if variablesStr != "" {
			if err := json.Unmarshal([]byte(variablesStr), &variables); err != nil {
				http.Error(w, "Invalid variables", http.StatusBadRequest)
				return
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute query
	result, err := h.ExecuteQuery(query, variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ExecuteQuery executes a GraphQL query
func (h *Handler) ExecuteQuery(query string, variables map[string]interface{}) (interface{}, error) {
	// Use graphql-go to execute the query
	params := graphql.Params{
		Schema:         h.GQLSchema,
		RequestString:  query,
		VariableValues: variables,
		Context:        context.Background(),
	}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("errors: %+v", result.Errors)
	}
	return result, nil
}
