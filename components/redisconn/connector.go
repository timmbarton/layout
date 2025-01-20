package redisconn

import (
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Host     string `validate:"required"`
	Port     uint16 `validate:"required"`
	Username string
	Password string
	DB       int
}

func Connect(cfg Config) (cl *redis.Client, err error) {
	cl = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err = cl.Ping(cl.Context()).Result()
	if err != nil {
		return cl, err
	}

	return cl, nil
}
