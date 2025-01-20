package redisconn

import (
	"context"
)

type Conn struct {
	c   *redis.Client
	cfg Config
}

func New(cfg Config) (*Conn, error) {
	return &Conn{
		c:   new(redis.Client),
		cfg: cfg,
	}, nil
}

func (c *Conn) Start(_ context.Context) error {
	conn, err := Connect(c.cfg)
	if err != nil {
		return err
	}

	*c.c = *conn

	return nil
}
func (c *Conn) Stop(_ context.Context) error {
	if c.c != nil {
		return c.c.Close()
	}

	return nil
}
func (c *Conn) GetName() string       { return "Redis" }
func (c *Conn) Client() *redis.Client { return c.c }
