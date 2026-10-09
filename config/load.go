package config

import (
	"os"
	mysqlAdapter "smsgateway/adapter/mysql"
	rabbitmqAdapter "smsgateway/adapter/rabbitmq"
	redisAdapter "smsgateway/adapter/redis"
	smsoperator "smsgateway/adapter/sms_operator"
	capacityRedis "smsgateway/capacity/redis"
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
		RabbitMQ: rabbitmqAdapter.Config{
			Username:      getEnv[string]("RABBITMQ_USERNAME", "guest"),
			Password:      getEnv[string]("RABBITMQ_PASSWORD", "guest"),
			Host:          getEnv[string]("RABBITMQ_HOST", "127.0.0.1"),
			Port:          getEnv[string]("RABBITMQ_PORT", "5672"),
			Vhost:         getEnv[string]("RABBITMQ_VHOST", "/"),
			Heartbeat:     getEnv[time.Duration]("RABBITMQ_HEARTBEAT", time.Second*10),
			DialTimeout:   getEnv[time.Duration]("RABBITMQ_DIAL_TIMEOUT", time.Second*30),
			PrefetchCount: getEnv[int]("RABBITMQ_PREFETCH_COUNT", 1),
		},
		Redis: redisAdapter.Config{
			Host:     getEnv[string]("REDIS_HOST", "127.0.0.1"),
			Port:     getEnv[string]("REDIS_PORT", "6379"),
			Password: getEnv[string]("REDIS_PASSWORD", ""),
			DB:       getEnv[int]("REDIS_DB", 0),
		},
		Capacity: capacityRedis.Config{
			KeyPrefix:    getEnv[string]("CAPACITY_KEY_PREFIX", "capacity"),
			SyncInterval: getEnv[time.Duration]("CAPACITY_SYNC_INTERVAL", time.Second*30),
			WaitTimeout:  getEnv[time.Duration]("CAPACITY_WAIT_TIMEOUT", time.Second*10),
			PollInterval: getEnv[time.Duration]("CAPACITY_POLL_INTERVAL", time.Millisecond*25),
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
