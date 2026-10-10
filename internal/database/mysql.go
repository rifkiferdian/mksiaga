package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"mksiaga/internal/config"
)

func Open(ctx context.Context, cfg config.Database) (*sql.DB, error) {
	dsn := mysql.NewConfig()
	dsn.User = cfg.User
	dsn.Passwd = cfg.Password
	dsn.Net = "tcp"
	dsn.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
	dsn.DBName = cfg.Name
	dsn.ParseTime = true
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, fmt.Errorf("load Jakarta timezone: %w", err)
	}
	dsn.Loc = jakarta
	dsn.Timeout = 5 * time.Second
	dsn.ReadTimeout = 5 * time.Second
	dsn.WriteTimeout = 5 * time.Second
	db, err := sql.Open("mysql", dsn.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect mysql: %w", err)
	}
	return db, nil
}
