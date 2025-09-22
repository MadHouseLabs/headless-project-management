package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Server    ServerConfig    `json:"server"`
	Database  DatabaseConfig  `json:"database"`
	Storage   StorageConfig   `json:"storage"`
	MCP       MCPConfig       `json:"mcp"`
	Embedding EmbeddingConfig `json:"embedding"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	DataDir string `json:"data_dir"`
}

type StorageConfig struct {
	UploadDir string `json:"upload_dir"`
}

type MCPConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type EmbeddingConfig struct {
	Provider       string `json:"provider"` // "azure_openai", "openai", "local"
	AzureEndpoint  string `json:"azure_endpoint"`
	AzureAPIKey    string `json:"azure_api_key"`
	DeploymentName string `json:"deployment_name"`
	Dimension      int    `json:"dimension"`
	Workers        int    `json:"workers"` // Number of embedding workers
}

func LoadConfig(path string) (*Config, error) {
	// Load from environment variables first, with defaults
	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			DataDir: getEnv("DATABASE_DIR", "./data"),
		},
		Storage: StorageConfig{
			UploadDir: getEnv("UPLOAD_DIR", "./data/uploads"),
		},
		MCP: MCPConfig{
			Enabled: getEnvAsBool("MCP_ENABLED", true),
			Port:    getEnvAsInt("MCP_PORT", 8081),
		},
		Embedding: EmbeddingConfig{
			Provider:       getEnv("EMBEDDING_PROVIDER", "azure_openai"),
			AzureEndpoint:  getEnv("AZURE_OPENAI_ENDPOINT", ""),
			AzureAPIKey:    getEnv("AZURE_OPENAI_API_KEY", ""),
			DeploymentName: getEnv("AZURE_OPENAI_EMBEDDING_DEPLOYMENT", "text-embedding-ada-002"),
			Dimension:      getEnvAsInt("EMBEDDING_DIMENSION", 1536),
			Workers:        getEnvAsInt("EMBEDDING_WORKERS", 3),
		},
	}

	// If a config file is specified, load it and override env vars
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			// File doesn't exist, use env vars only
		} else {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(config); err != nil {
				return nil, err
			}
		}
	}

	if !filepath.IsAbs(config.Database.DataDir) {
		config.Database.DataDir, _ = filepath.Abs(config.Database.DataDir)
	}
	if !filepath.IsAbs(config.Storage.UploadDir) {
		config.Storage.UploadDir, _ = filepath.Abs(config.Storage.UploadDir)
	}

	return config, nil
}


// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}