package config

import (
    "log"
    "os"
    "reflect"
    "strconv"
    "strings"

    "github.com/BurntSushi/toml"
)

type Config struct {
    Auth     AuthConfig     `toml:"auth"`
    Database DatabaseConfig `toml:"database"`
    Redis    RedisConfig    `toml:"redis"`
    Worker   WorkerConfig   `toml:"worker"`
}

type AuthConfig struct {
    Username string `toml:"username"`
    Password string `toml:"password"`
}

type DatabaseConfig struct {
    Host     string `toml:"host"`
    Port     int    `toml:"port"`
    User     string `toml:"user"`
    Password string `toml:"password"`
    Name     string `toml:"name"`
    SSLMode  string `toml:"sslmode"`
}

type RedisConfig struct {
    Host string `toml:"host"`
    Port int    `toml:"port"`
}

type WorkerConfig struct {
    MaxAsyncProcess int `toml:"max_async_process"`
}

// LoadDefaultConfig loads default config from /config/default.toml
func LoadDefaultConfig() (*Config, error) {
    var config Config
    if _, err := toml.DecodeFile("config/default.toml", &config); err != nil {
        return nil, err
    }
    return &config, nil
}

// LoadConfig loads config from default TOML and overrides from environment variables
func LoadConfig() (*Config, error) {
    cfg, err := LoadDefaultConfig()
    if err != nil {
        return nil, err
    }

    UpdateEnvConfig(cfg)
    return cfg, nil
}

// UpdateEnvConfig updates config fields by checking environment variables
func UpdateEnvConfig(cfg *Config) {
    // Use reflection to iterate through fields and update from env vars
    v := reflect.ValueOf(cfg).Elem()
    t := v.Type()

    // Helper function to set fields if env matches toml tag (uppercased and underscores)
    for i := 0; i < v.NumField(); i++ {
        structField := t.Field(i)
        sectionVal := v.Field(i)
        if sectionVal.Kind() == reflect.Struct {
            updateSectionFromEnv(sectionVal, structField.Name)
        }
    }
}

func updateSectionFromEnv(val reflect.Value, sectionName string) {
    t := val.Type()
    sectionPrefix := strings.ToUpper(sectionName) + "_"

    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldType := t.Field(i)
        tag := fieldType.Tag.Get("toml")
        envKey := sectionPrefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))

        envVal, exists := os.LookupEnv(envKey)
        if !exists {
            continue
        }

        if !field.CanSet() {
            continue
        }

        switch field.Kind() {
        case reflect.String:
            field.SetString(envVal)
        case reflect.Int:
            if intVal, err := strconv.Atoi(envVal); err == nil {
                field.SetInt(int64(intVal))
            } else {
                log.Printf("Warning: invalid integer for env %s\n", envKey)
            }
        // Add cases if you expand structs with bool, float, etc.
        }
    }
}
