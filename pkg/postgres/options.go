package postgres

import (
	"fmt"
	"time"

	"github.com/mathbdw/book/config"
)

// Option -.
type Option func(*Postgres)

// Driver - Set driver
func Driver(driver string) Option {
	return func(p *Postgres) {
		p.driver = driver
	}
}

// Dsn - Set data source name
func Dsn(cfg config.Database) Option {
	return func(p *Postgres) {
		p.dsn = fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v",
			cfg.Host,
			cfg.Port,
			cfg.User,
			cfg.Password,
			cfg.Name,
			cfg.SslMode,
		)
	}
}

// MaxOpenConns - Set maximum open connections
func MaxOpenConns(cnt int) Option {
	return func(p *Postgres) {
		p.maxOpenConns = cnt
	}
}

// MaxIdleConns - Set maximum Idle connections
func MaxIdleConns(cnt int) Option {
	return func(p *Postgres) {
		p.maxIdleConns = cnt
	}
}

// ConnMaxIdleTime - Set maximum connection idle time
func ConnMaxIdleTime(time time.Duration) Option {
	return func(p *Postgres) {
		p.connMaxIdleTime = time
	}
}

// ConnMaxLifetime - Set maximum connection lifetime
func ConnMaxLifetime(time time.Duration) Option {
	return func(p *Postgres) {
		p.connMaxLifeTime = time
	}
}
