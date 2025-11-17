package config

import (
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Env       string          `toml:"env"`
	Auth      AuthConfig      `toml:"auth"`
	Scheduler SchedulerConfig `toml:"scheduler"`
	Database  DatabaseConfig  `toml:"database"`
	Redis     RedisConfig     `toml:"redis"`
	Worker    WorkerConfig    `toml:"worker"`
	Providers ProvidersConfig `toml:"providers"`
	RateLimit RateLimitConfig `toml:"rate_limit"`
}

type AuthConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type SchedulerConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Enabled  bool   `toml:"enabled"`
}

type DatabaseConfig struct {
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	User        string `toml:"user"`
	Password    string `toml:"password"`
	Name        string `toml:"name"`
	SSLMode     string `toml:"sslmode"`
	AutoMigrate bool   `toml:"auto_migrate"`
}

type RedisConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type WorkerConfig struct {
	MaxAsyncProcess int `toml:"max_async_process"`
}

type RateLimitConfig struct {
	Enabled     bool                     `toml:"enabled"`
	Storage     string                   `toml:"storage"` // "redis" or "memory"
	DefaultRule RateLimitRule            `toml:"default"`
	Routes      map[string]RateLimitRule `toml:"routes"`
}

type RateLimitRule struct {
	BucketSize     int           `toml:"bucket_size"`     // Maximum tokens in bucket
	RefillSize     int           `toml:"refill_size"`     // Tokens added per refill
	RefillDuration time.Duration `toml:"refill_duration"` // How often to refill
	IdentifyBy     string        `toml:"identify_by"`     // "ip" or "api_key"
	Enabled        bool          `toml:"enabled"`
}

type ProvidersConfig struct {
	Enabled       []string                      `toml:"enabled"`
	LoadBalancing string                        `toml:"load_balancing"` // "round_robin", "weighted", "least_load"
	SMTP          map[string]SMTPProviderConfig `toml:"smtp"`
	API           map[string]APIProviderConfig  `toml:"api"`
}

type SMTPConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	From     string `toml:"from"`
}

type SMTPProviderConfig struct {
	Host             string `toml:"host"`
	Port             int    `toml:"port"`
	Username         string `toml:"username"`
	Password         string `toml:"password"`
	From             string `toml:"from"`
	Priority         int    `toml:"priority"`
	MaxEmailsPerHour int    `toml:"max_emails_per_hour"`
}

type APIProviderConfig struct {
	Endpoint         string `toml:"endpoint"`
	Token            string `toml:"token"`
	From             string `toml:"from"`
	Priority         int    `toml:"priority"`
	MaxEmailsPerHour int    `toml:"max_emails_per_hour"`
	BulkEnabled      bool   `toml:"bulk_enabled"`
	MaxBatchSize     int    `toml:"max_batch_size"`
}

// LoadDefaultConfig loads default config from env/default.toml
func LoadDefaultConfig() (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile("env/default.toml", &config); err != nil {
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
		} else {
			// Handle top-level fields
			tag := structField.Tag.Get("toml")
			envKey := strings.ToUpper(tag)
			if envVal, exists := os.LookupEnv(envKey); exists {
				updateFieldValue(sectionVal, envVal)
			}
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

		updateFieldValue(field, envVal)
	}
}

func updateFieldValue(field reflect.Value, envVal string) {
	if !field.CanSet() {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envVal)
	case reflect.Int:
		if intVal, err := strconv.Atoi(envVal); err == nil {
			field.SetInt(int64(intVal))
		} else {
			log.Printf("Warning: invalid integer for env value: %s", envVal)
		}
	case reflect.Bool:
		if boolVal, err := strconv.ParseBool(envVal); err == nil {
			field.SetBool(boolVal)
		} else {
			log.Printf("Warning: invalid boolean for env value: %s", envVal)
		}
	case reflect.TypeOf(time.Duration(0)).Kind():
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			if duration, err := time.ParseDuration(envVal); err == nil {
				field.Set(reflect.ValueOf(duration))
			} else {
				log.Printf("Warning: invalid duration for env value: %s", envVal)
			}
		}
	}
}

// SendGridConfig represents SendGrid email provider configuration
type SendGridConfig struct {
	APIKey           string `toml:"api_key"`
	From             string `toml:"from"`
	Enabled          bool   `toml:"enabled"`
	Priority         int    `toml:"priority"`
	MaxEmailsPerHour int    `toml:"max_emails_per_hour"`
	MaxBatchSize     int    `toml:"max_batch_size"`
	BulkEnabled      bool   `toml:"bulk_enabled"`
}

// MailtrapConfig represents Mailtrap email provider configuration
type MailtrapConfig struct {
	APIToken         string `toml:"api_token"`
	BulkEndpoint     string `toml:"bulk_endpoint"`
	From             string `toml:"from"`
	Enabled          bool   `toml:"enabled"`
	Priority         int    `toml:"priority"`
	MaxEmailsPerHour int    `toml:"max_emails_per_hour"`
	MaxBatchSize     int    `toml:"max_batch_size"`
	BulkEnabled      bool   `toml:"bulk_enabled"`
}
