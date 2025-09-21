package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                    string
	DBPort                    string
	DBUser                    string
	DBPassword                string
	DBName                    string
	DBSSLMode                 string
	ClickEventTimeWindowHours int
	REDISURL                  string
	REDISPassword             string
	REDISDBStr                string
	KafkaUrl                  string
	KafkaClickTopic           string
	KafkaConversionTopic      string
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	timeWindowHours := 24
	if timeWindowStr := os.Getenv("CLICK_EVENT_TIME_WINDOW_HOURS"); timeWindowStr != "" {
		if parsed, err := strconv.Atoi(timeWindowStr); err == nil {
			timeWindowHours = parsed
		}
	}

	return &Config{
		DBHost:                    getEnv("DB_HOST", "localhost"),
		DBPort:                    getEnv("DB_PORT", "5432"),
		DBUser:                    getEnv("DB_USER", "postgres"),
		DBPassword:                getEnv("DB_PASSWORD", ""),
		DBName:                    getEnv("DB_NAME", "tyrattribution"),
		DBSSLMode:                 getEnv("DB_SSL_MODE", "disable"),
		ClickEventTimeWindowHours: timeWindowHours,
		REDISURL:                  getEnv("REDIS_URL", "redis:6379"),
		REDISPassword:             getEnv("REDIS_PASSWORD", ""),
		REDISDBStr:                getEnv("REDIS_DB", "0"),
		KafkaUrl:                  getEnv("KAFKA_BROKER_URL", "kafka:9092"),
		KafkaClickTopic:           getEnv("KAFKA_CLICK_EVENT_TOPIC", "click_event"),
		KafkaConversionTopic:      getEnv("KAFKA_CONVERSION_EVENT_TOPIC", "click_conversion"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
