package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost                     string
	DBPort                     string
	DBUser                     string
	DBPassword                 string
	DBName                     string
	ClickEventTimeWindowHours  int
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
		ClickEventTimeWindowHours: timeWindowHours,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}