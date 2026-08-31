package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port                     string
	Env                      string
	LogLevel                 string
	PostgresHost             string
	PostgresPort             string
	PostgresDB               string
	PostgresUser             string
	PostgresPassword         string
	PostgresSSLMode          string
	RedisHost                string
	RedisPort                string
	RedisPassword            string
	KafkaBrokers             string
	KafkaEnabled             bool
	KafkaGroupID             string
	MLServiceURL             string
	MLServiceTimeoutSeconds  int
	RazorpayKeyID            string
	RazorpayKeySecret        string
}

func LoadConfig() *Config {
		port := getEnv("PORT", "8082")
	env := getEnv("ENV", "development")
	logLevel := getEnv("LOG_LEVEL", "info")

	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgDB := getEnv("POSTGRES_DB", "rri_db")
	pgUser := getEnv("POSTGRES_USER", "rri_user")
	pgPass := getEnv("POSTGRES_PASSWORD", "rri_secure_password")
	pgSSL := getEnv("POSTGRES_SSLMODE", "disable")

	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPass := getEnv("REDIS_PASSWORD", "")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaEnabledStr := getEnv("KAFKA_ENABLED", "false")
	kafkaEnabled := kafkaEnabledStr == "true" || kafkaEnabledStr == "1"
	kafkaGroupID := getEnv("KAFKA_GROUP_ID", "rri-recovery-group")

	mlServiceURL := getEnv("ML_SERVICE_URL", "http://localhost:8000")
	mlTimeout, _ := strconv.Atoi(getEnv("ML_SERVICE_TIMEOUT_SECONDS", "5"))

	rzpKeyID := getEnv("RAZORPAY_KEY_ID", "rzp_test_rri_recovery_sim")
	rzpKeySecret := getEnv("RAZORPAY_KEY_SECRET", "rzp_secret_sim_mock_only")

	return &Config{
		Port:                    port,
		Env:                     env,
		LogLevel:                logLevel,
		PostgresHost:            pgHost,
		PostgresPort:            pgPort,
		PostgresDB:              pgDB,
		PostgresUser:            pgUser,
		PostgresPassword:        pgPass,
		PostgresSSLMode:         pgSSL,
		RedisHost:               redisHost,
		RedisPort:               redisPort,
		RedisPassword:           redisPass,
		KafkaBrokers:            kafkaBrokers,
		KafkaEnabled:            kafkaEnabled,
		KafkaGroupID:            kafkaGroupID,
		MLServiceURL:            mlServiceURL,
		MLServiceTimeoutSeconds: mlTimeout,
		RazorpayKeyID:           rzpKeyID,
		RazorpayKeySecret:       rzpKeySecret,
	}
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser, c.PostgresPassword, c.PostgresHost, c.PostgresPort, c.PostgresDB, c.PostgresSSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
