package postgresconn

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Conn struct {
	db  *sqlx.DB
	cfg Config
}

func New(cfg Config) (*Conn, error) {
	return &Conn{
		db:  new(sqlx.DB),
		cfg: cfg,
	}, nil
}

func (c *Conn) Start(_ context.Context) error {
	conn, err := Connect(c.cfg)
	if err != nil {
		return err
	}

	*c.db = *conn

	return nil
}
func (c *Conn) Stop(_ context.Context) error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
}
func (c *Conn) GetName() string { return "Postgres" }
func (c *Conn) DB() *sqlx.DB    { return c.db }
