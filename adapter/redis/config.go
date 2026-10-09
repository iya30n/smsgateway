package redis

import "fmt"

type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
