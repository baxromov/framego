package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"framego/pkg/orm"
)

// Config represents the application configuration
type Config struct {
	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Server configuration
	Server ServerConfig `json:"server"`

	// Debug mode
	Debug bool `json:"debug"`

	// Secret key for signing tokens
	SecretKey string `json:"secret_key"`

	// GraphQL configuration
	GraphQL GraphQLConfig `json:"graphql"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// GraphQLConfig represents the GraphQL configuration
type GraphQLConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Driver:   "sqlite3",
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "",
			Database: "test.db",
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Debug:     true,
		SecretKey: generateRandomKey(32),
		GraphQL: GraphQLConfig{
			Enabled: false,
			Path:    "/graphql",
		},
	}
}

// LoadFromFile loads configuration from a file
func LoadFromFile(filename string) (*Config, error) {
	// Start with default config
	config := DefaultConfig()

	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return config, nil
}

// SaveToFile saves configuration to a file
func (c *Config) SaveToFile(filename string) error {
	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(filename))
	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Write file
	if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToORMConfig converts the database configuration to an ORM configuration
func (c *Config) ToORMConfig() orm.Config {
	return orm.Config{
		Driver:   c.Database.Driver,
		Host:     c.Database.Host,
		Port:     c.Database.Port,
		User:     c.Database.User,
		Password: c.Database.Password,
		Database: c.Database.Database,
	}
}

// generateRandomKey generates a random key of the specified length
func generateRandomKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		// Use crypto/rand for secure random number generation
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			// Fallback to less secure but functional method if crypto/rand fails
			result[i] = charset[time.Now().Nanosecond()%len(charset)]
			continue
		}
		result[i] = charset[int(b[0])%len(charset)]
	}
	return string(result)
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := DefaultConfig()

	// Database configuration
	config.Database.Driver = GetEnv("DB_DRIVER", config.Database.Driver)
	config.Database.Host = GetEnv("DB_HOST", config.Database.Host)
	if port := GetEnv("DB_PORT", ""); port != "" {
		fmt.Sscanf(port, "%d", &config.Database.Port)
	}
	config.Database.User = GetEnv("DB_USER", config.Database.User)
	config.Database.Password = GetEnv("DB_PASSWORD", config.Database.Password)
	config.Database.Database = GetEnv("DB_NAME", config.Database.Database)

	// Server configuration
	config.Server.Host = GetEnv("SERVER_HOST", config.Server.Host)
	if port := GetEnv("SERVER_PORT", ""); port != "" {
		fmt.Sscanf(port, "%d", &config.Server.Port)
	}

	// Debug mode
	if debug := GetEnv("DEBUG", ""); debug != "" {
		config.Debug = debug == "true" || debug == "1"
	}

	// Secret key
	config.SecretKey = GetEnv("SECRET_KEY", config.SecretKey)

	// GraphQL configuration
	if enabled := GetEnv("GRAPHQL_ENABLED", ""); enabled != "" {
		config.GraphQL.Enabled = enabled == "true" || enabled == "1"
	}
	config.GraphQL.Path = GetEnv("GRAPHQL_PATH", config.GraphQL.Path)

	return config
}
