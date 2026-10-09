package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Adapter struct {
	config Config
	client *redis.Client
}

func New(config Config) (*Adapter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr(),
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Adapter{config: config, client: client}, nil
}

func (a *Adapter) Client() *redis.Client {
	return a.client
}

func (a *Adapter) Close() error {
	if a.client == nil {
		return nil
	}

	return a.client.Close()
}
