package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"framego/pkg/models"
)

// ORM represents the object-relational mapper
type ORM struct {
	db        *sql.DB
	driver    string
	models    map[string]models.ModelInterface
	connected bool
}

// Config represents the configuration for the ORM
type Config struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// New creates a new ORM instance
func New(config Config) (*ORM, error) {
	dsn := buildDSN(config)
	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &ORM{
		db:        db,
		driver:    config.Driver,
		models:    make(map[string]models.ModelInterface),
		connected: true,
	}, nil
}

// buildDSN builds the data source name for the database connection
func buildDSN(config Config) string {
	switch config.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			config.User, config.Password, config.Host, config.Port, config.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.Database)
	case "sqlite3":
		return config.Database
	default:
		return ""
	}
}

// Close closes the database connection
func (o *ORM) Close() error {
	if o.db != nil {
		return o.db.Close()
	}
	return nil
}

// RegisterModel registers a model with the ORM
func (o *ORM) RegisterModel(model models.ModelInterface) error {
	if err := model.Validate(); err != nil {
		return err
	}

	o.models[model.GetTableName()] = model
	return nil
}

// CreateTables creates tables for all registered models
func (o *ORM) CreateTables() error {
	for _, model := range o.models {
		if err := o.createTable(model); err != nil {
			return err
		}
	}
	return nil
}

// createTable creates a table for the given model
func (o *ORM) createTable(model models.ModelInterface) error {
	tableName := model.GetTableName()
	fields := model.GetFields()

	var columns []string
	var primaryKeys []string
	var foreignKeys []string

	for name, field := range fields {
		column := fmt.Sprintf("%s %s", name, getSQLType(field, o.driver))

		if field.NotNull {
			column += " NOT NULL"
		}

		if field.Unique {
			column += " UNIQUE"
		}

		if field.Default != nil {
			column += fmt.Sprintf(" DEFAULT %v", field.Default)
		}

		if field.PrimaryKey {
			primaryKeys = append(primaryKeys, name)
		}

		if field.AutoIncrement && o.driver == "mysql" {
			column += " AUTO_INCREMENT"
		}

		columns = append(columns, column)

		if field.ForeignKey != nil {
			fk := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
				name, field.ForeignKey.Model, field.ForeignKey.Field)

			if field.ForeignKey.OnDelete != "" {
				fk += fmt.Sprintf(" ON DELETE %s", field.ForeignKey.OnDelete)
			}

			if field.ForeignKey.OnUpdate != "" {
				fk += fmt.Sprintf(" ON UPDATE %s", field.ForeignKey.OnUpdate)
			}

			foreignKeys = append(foreignKeys, fk)
		}
	}

	if len(primaryKeys) > 0 {
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	columns = append(columns, foreignKeys...)

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
		tableName, strings.Join(columns, ", "))

	_, err := o.db.Exec(query)
	return err
}

// getSQLType returns the SQL type for the given field
func getSQLType(field models.Field, driver string) string {
	switch field.Type.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "INTEGER"
	case reflect.Int64, reflect.Uint64:
		return "BIGINT"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.String:
		if field.MaxLength > 0 {
			return fmt.Sprintf("VARCHAR(%d)", field.MaxLength)
		}
		return "TEXT"
	default:
		if field.Type == reflect.TypeOf(time.Time{}) {
			if driver == "mysql" {
				return "DATETIME"
			}
			return "TIMESTAMP"
		}
		return "TEXT"
	}
}

// Create inserts a new record into the database
func (o *ORM) Create(tableName string, data map[string]interface{}) (int64, error) {
	model, ok := o.models[tableName]
	if !ok {
		return 0, fmt.Errorf("model %s not registered", tableName)
	}

	fields := model.GetFields()
	var columns []string
	var placeholders []string
	var values []interface{}

	i := 1
	for column, value := range data {
		if _, ok := fields[column]; !ok {
			return 0, fmt.Errorf("field %s not found in model %s", column, tableName)
		}

		columns = append(columns, column)

		if o.driver == "postgres" {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		} else {
			placeholders = append(placeholders, "?")
		}

		values = append(values, value)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := o.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Get retrieves a record from the database
func (o *ORM) Get(tableName string, id interface{}) (map[string]interface{}, error) {
	model, ok := o.models[tableName]
	if !ok {
		return nil, fmt.Errorf("model %s not registered", tableName)
	}

	fields := model.GetFields()
	var primaryKey string
	for name, field := range fields {
		if field.PrimaryKey {
			primaryKey = name
			break
		}
	}

	if primaryKey == "" {
		return nil, fmt.Errorf("model %s has no primary key", tableName)
	}

	var placeholder string
	if o.driver == "postgres" {
		placeholder = "$1"
	} else {
		placeholder = "?"
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = %s", tableName, primaryKey, placeholder)

	row := o.db.QueryRow(query, id)

	columns, err := getColumns(tableName, o.db)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, column := range columns {
		result[column] = values[i]
	}

	return result, nil
}

// getColumns returns the column names for the given table
func getColumns(tableName string, db *sql.DB) ([]string, error) {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 1", tableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	return columns, nil
}

// Update updates a record in the database
func (o *ORM) Update(tableName string, id interface{}, data map[string]interface{}) error {
	model, ok := o.models[tableName]
	if !ok {
		return fmt.Errorf("model %s not registered", tableName)
	}

	fields := model.GetFields()
	var primaryKey string
	for name, field := range fields {
		if field.PrimaryKey {
			primaryKey = name
			break
		}
	}

	if primaryKey == "" {
		return fmt.Errorf("model %s has no primary key", tableName)
	}

	var setStatements []string
	var values []interface{}

	i := 1
	for column, value := range data {
		if _, ok := fields[column]; !ok {
			return fmt.Errorf("field %s not found in model %s", column, tableName)
		}

		if column == primaryKey {
			continue
		}

		var placeholder string
		if o.driver == "postgres" {
			placeholder = fmt.Sprintf("$%d", i)
		} else {
			placeholder = "?"
		}

		setStatements = append(setStatements, fmt.Sprintf("%s = %s", column, placeholder))
		values = append(values, value)
		i++
	}

	var placeholder string
	if o.driver == "postgres" {
		placeholder = fmt.Sprintf("$%d", i)
	} else {
		placeholder = "?"
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s",
		tableName, strings.Join(setStatements, ", "), primaryKey, placeholder)

	values = append(values, id)

	_, err := o.db.Exec(query, values...)
	return err
}

// Delete deletes a record from the database
func (o *ORM) Delete(tableName string, id interface{}) error {
	model, ok := o.models[tableName]
	if !ok {
		return fmt.Errorf("model %s not registered", tableName)
	}

	fields := model.GetFields()
	var primaryKey string
	for name, field := range fields {
		if field.PrimaryKey {
			primaryKey = name
			break
		}
	}

	if primaryKey == "" {
		return fmt.Errorf("model %s has no primary key", tableName)
	}

	var placeholder string
	if o.driver == "postgres" {
		placeholder = "$1"
	} else {
		placeholder = "?"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", tableName, primaryKey, placeholder)

	_, err := o.db.Exec(query, id)
	return err
}

// Query executes a custom query and returns the results
func (o *ORM) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		result := make(map[string]interface{})
		for i, column := range columns {
			result[column] = values[i]
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
