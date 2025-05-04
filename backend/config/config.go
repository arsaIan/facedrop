package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerConfig   ServerConfig
	DatabaseConfig DatabaseConfig
	JWTConfig      JWTConfig
	StorageConfig  StorageConfig
	DeepFaceConfig DeepFaceConfig
	EmailConfig    EmailConfig 
}

type ServerConfig struct {
	Port        int
	Environment string
	BaseURL        string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	Migrate  bool
}

type JWTConfig struct {
	Secret    string
	Expiry    time.Duration
}

type StorageConfig struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	EventBucket		string
	UserBucket 		string
	ZipBucket		string
}

type DeepFaceConfig struct {
	URL string
}

type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	From         string `yaml:"from"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	jwtExpiry, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY"))

	return &Config{
		ServerConfig: ServerConfig{
			Port:        port,
			Environment: os.Getenv("ENVIRONMENT"),
			BaseURL:     os.Getenv("BASE_URL"),
		},
		DatabaseConfig: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     dbPort,
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
			Migrate:  os.Getenv("DB_MIGRATE") == "true",
		},
		JWTConfig: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
			Expiry: time.Duration(jwtExpiry) * time.Hour,
		},
		StorageConfig: StorageConfig{
			Endpoint:        os.Getenv("STORAGE_ENDPOINT"),
			Region:          os.Getenv("STORAGE_REGION"),
			AccessKeyID:     os.Getenv("STORAGE_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("STORAGE_SECRET_ACCESS_KEY"),
			EventBucket:     os.Getenv("EVENT_BUCKET"),
			UserBucket:      os.Getenv("USER_BUCKET"),
			ZipBucket:       os.Getenv("ZIP_BUCKET"),
		},
		DeepFaceConfig: DeepFaceConfig{
			URL: os.Getenv("DEEP_FACE_URL"),
		},
		EmailConfig: EmailConfig{
			SMTPHost:     os.Getenv("EMAIL_SMTP_HOST"),
			SMTPPort:     getEnvAsInt("EMAIL_SMTP_PORT", 587),
			Username:     os.Getenv("EMAIL_USERNAME"),
			Password:     os.Getenv("EMAIL_PASSWORD"),
			From:         os.Getenv("EMAIL_FROM"),
		},
	}, nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DatabaseConfig.Host,
		c.DatabaseConfig.Port,
		c.DatabaseConfig.User,
		c.DatabaseConfig.Password,
		c.DatabaseConfig.Name,
		c.DatabaseConfig.SSLMode,
	)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
} 