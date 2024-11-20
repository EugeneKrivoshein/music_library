package conn

import (
	"database/sql"
	"fmt"

	"github.com/EugeneKrivoshein/music_library/config"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type PostgresProvider struct {
	db *sql.DB
}

func (p *PostgresProvider) Close() error {
	return p.db.Close()
}

func (p *PostgresProvider) DB() *sql.DB {
	return p.db
}

func NewPostgresProvider(cfg *config.Config) (*PostgresProvider, error) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Errorf("Ошибка подключения к базе данных: %v", err)
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Errorf("Ошибка при проверке подключения: %v", err)
		return nil, fmt.Errorf("база данных недоступна: %w", err)
	}

	return &PostgresProvider{db: db}, nil
}
