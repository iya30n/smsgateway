package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	return Config{
		
	}
}

func getEnv[T string | int | time.Duration](key string, defaultValue T) T {
	var value any
	value = os.Getenv(key)
	if value.(string) == "" {
		return defaultValue
	}

	switch any(defaultValue).(type) {
	case string:
		return value.(T)
	case int:
		val, err := strconv.Atoi(value.(string))
		if err != nil {
			panic(err)
		}

		return any(val).(T)
	case time.Duration:
		t, err := time.ParseDuration(value.(string))
		if err != nil {
			panic(err)
		}

		return any(t).(T)
	}

	return value.(T)
}
