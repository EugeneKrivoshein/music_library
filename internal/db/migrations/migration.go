package migrations

import (
	"fmt"

	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func RunMigrations(provider *conn.PostgresProvider) error {
	log := logrus.New()

	db := provider.DB()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS songs (
			id SERIAL PRIMARY KEY,
			group_name VARCHAR(255) NOT NULL,
			song_name VARCHAR(255) NOT NULL,
			release_date DATE,
			text TEXT,
			link TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		log.Printf("Выполнение миграции: %s", query)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка выполнения миграции: %w", err)
		}
	}
	log.Info("Миграции успешно выполнены.")
	return nil
}
