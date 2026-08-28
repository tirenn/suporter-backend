package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port               string `mapstructure:"PORT"`
	DBHost             string `mapstructure:"DB_HOST"`
	DBPort             string `mapstructure:"DB_PORT"`
	DBUser             string `mapstructure:"DB_USER"`
	DBPassword         string `mapstructure:"DB_PASSWORD"`
	DBName             string `mapstructure:"DB_NAME"`
	DBSSLMode          string `mapstructure:"DB_SSLMODE"`
	JWTSecret          string `mapstructure:"JWT_SECRET"`
	JWTExpiryHr        int    `mapstructure:"JWT_EXPIRY_HOURS"`
	RecaptchaSecretKey string `mapstructure:"RECAPTCHA_SECRET_KEY"`
	EnableWebhookHMAC  bool   `mapstructure:"ENABLE_WEBHOOK_HMAC"`
	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AppURL             string `mapstructure:"APP_URL"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_URL", "http://localhost:8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "suporter")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("RECAPTCHA_SECRET_KEY", "6LeIxAcTAAAAAlO_PuGNMaximum-T6Nmcca_WDm")
	viper.SetDefault("ENABLE_WEBHOOK_HMAC", true)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "*")

	viper.BindEnv("PORT")
	viper.BindEnv("APP_URL")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("DB_SSLMODE")
	viper.BindEnv("JWT_SECRET")
	viper.BindEnv("JWT_EXPIRY_HOURS")
	viper.BindEnv("RECAPTCHA_SECRET_KEY")
	viper.BindEnv("ENABLE_WEBHOOK_HMAC")
	viper.BindEnv("CORS_ALLOWED_ORIGINS")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("[Viper Notice] Could not read .env file (%v). Using environment variables.", err)
	} else {
		log.Printf("[Viper Config] Loaded environment configuration from .env")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[Viper Error] Failed to unmarshal configuration: %v", err)
	}

	// Fallback to direct environment variables if Viper unmarshal didn't populate
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = viper.GetString("JWT_SECRET")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = os.Getenv("JWT_SECRET")
	}

	if cfg.JWTSecret == "" {
		log.Fatalf("[Config Error] JWT_SECRET environment variable is required and cannot be empty")
	}

	return &cfg
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) RootDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBSSLMode)
}
