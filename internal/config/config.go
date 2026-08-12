package config

import (
	"fmt"
	"log"

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
	WebhookSecret      string `mapstructure:"WEBHOOK_SECRET"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Set fallback defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "suporter")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET", "suporter-super-secret-jwt-key-2026")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	// Google reCAPTCHA v2 — replace with your real secret from console.cloud.google.com/recaptcha
	viper.SetDefault("RECAPTCHA_SECRET_KEY", "6LeIxAcTAAAAAlO_PuGNMaximum-T6Nmcca_WDm")
	viper.SetDefault("WEBHOOK_SECRET", "suporter-webhook-hmac-secret-2026")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("[Viper Notice] Could not read .env file (%v). Using environment defaults.", err)
	} else {
		log.Printf("[Viper Config] Loaded environment configuration from .env")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("[Viper Error] Failed to unmarshal configuration: %v", err)
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
