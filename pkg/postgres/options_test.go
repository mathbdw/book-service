package postgres

import (
	"fmt"
	"testing"
	"time"

	"github.com/mathbdw/book/config"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	pg := &Postgres{}
	opt := Driver("driver")
	opt(pg)

	assert.Equal(t, "driver", pg.driver)
}

func TestDsn(t *testing.T) {
	db := config.Database{
		Host:     "test",
		Port:     1,
		User:     "user",
		Password: "password",
		Name:     "db",
		SslMode:  "disable",
	}
	dsnEx := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
		db.SslMode,
	)

	pg := &Postgres{}
	opt := Dsn(db)
	opt(pg)

	assert.Equal(t, dsnEx, pg.dsn)
}

func TestMaxOpenConns(t *testing.T) {
	pg := &Postgres{}
	opt := MaxOpenConns(1)
	opt(pg)

	assert.Equal(t, 1, pg.maxOpenConns)
}

func TestMaxIdleConns(t *testing.T) {
	pg := &Postgres{}
	opt := MaxIdleConns(1)
	opt(pg)

	assert.Equal(t, 1, pg.maxIdleConns)
}

func TestConnMaxIdleTime(t *testing.T) {
	timeSecond := time.Second
	pg := &Postgres{}
	opt := ConnMaxIdleTime(timeSecond)
	opt(pg)

	assert.Equal(t, timeSecond, pg.connMaxIdleTime)
}

func TestConnMaxLifetime(t *testing.T) {
	timeSecond := time.Second
	pg := &Postgres{}
	opt := ConnMaxLifetime(timeSecond)
	opt(pg)

	assert.Equal(t, timeSecond, pg.connMaxLifeTime)
}
