package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	AdminToken  string            `json:"adminToken"`
	CacheConfig CacheConfig       `json:"cache"`
	Log         LogConfig         `json:"log"`
	FileStorage FileStorageConfig `json:"fileStorage"`
}

type ServerConfig struct {
	Address string `json:"address"`
}

type LogConfig struct {
	Level string `json:"level"`
}

type CacheConfig struct {
	TTL        int `json:"ttl"`
	MaxEntries int `json:"max_entries"`
}

type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Name            string `json:"name"`
	SSLMode         string `json:"sslmode"`
	MaxOpenConns    int    `json:"maxOpenConns"`
	MaxIdleConns    int    `json:"maxIdleConns"`
	ConnMaxLifetime int    `json:"connMaxLifetime"`
}

type FileStorageConfig struct {
	Path string `json:"path"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
