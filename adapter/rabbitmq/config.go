package rabbitmq

import (
	"fmt"
	"time"
)

type Config struct {
	Username      string
	Password      string
	Host          string
	Port          string
	Vhost         string
	Heartbeat     time.Duration
	DialTimeout   time.Duration
	PrefetchCount int
}

func (c Config) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s", c.Username, c.Password, c.Host, c.Port)
}
