package postgresconn

import (
	"crypto/tls"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	Host           string      `validate:"required"`
	Port           uint16      `validate:"required"`
	Database       string      `validate:"required"`
	User           string      `validate:"required"`
	Password       string      `validate:"required"`
	TLSConfig      *tls.Config // nil disables TLS
	ConnectTimeout int         `validate:"required"` // seconds
}

func (c *Config) String() string {
	sslMode := "disable"
	sslConfig := ""
	if c.TLSConfig != nil {
		sslMode = "enable"
		// TODO sslConfig generation
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s %s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		sslMode,
		sslConfig,
	)
}

func Connect(cfg Config) (*sqlx.DB, error) {
	return sqlx.Connect("postgres", cfg.String())
}
