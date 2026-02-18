package config

import (
	"os"
)

type Config struct {
	GCSBucketName string
	GCSProjectID  string
	OpenAIAPIKey  string
	Port          string
}

func Load() *Config {
	return &Config{
		GCSBucketName: getEnv("GCS_BUCKET_NAME", "holodeck-art-submissions"),
		GCSProjectID:  getEnv("GCS_PROJECT_ID", ""),
		OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),
		Port:          getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}