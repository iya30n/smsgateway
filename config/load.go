package config

import (
	"os"
	mysqlAdapter "smsgateway/adapter/mysql"
	smsoperator "smsgateway/adapter/sms_operator"
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
		HttpServer: HttpServer{
			Host: getEnv[string]("HTTP_HOST", "127.0.0.1"),
			Port: getEnv[string]("HTTP_PORT", "8080"),
		},
		Mysql: mysqlAdapter.Config{
			Username: getEnv[string]("DB_USERNAME", "root"),
			Password: getEnv[string]("DB_PASSWORD", "smsgw@1234"),
			Host:     getEnv[string]("DB_HOST", "127.0.0.1"),
			Port:     getEnv[string]("DB_PORT", "3306"),
			DBName:   getEnv[string]("DB_NAME", "smsgateway"),
		},
		FirstOperator: smsoperator.Config{
			BaseURL: getEnv[string]("FIRST_OPERATOR_BASE_URL", "https://api.first-operator.ir"),
			APIKey:  getEnv[string]("FIRST_OPERATOR_API_KEY", ""),
			Timeout: getEnv[time.Duration]("FIRST_OPERATOR_TIMEOUT", time.Second*10),
		},
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
