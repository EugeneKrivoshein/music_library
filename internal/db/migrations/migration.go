package migrations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/sirupsen/logrus"
)

type Migration struct {
	Version string
	Up      string
	Down    string
}

func getMigrations(path string) ([]Migration, error) {
	var migrations []Migration
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать директорию миграций: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".up.sql" {
			version := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))-3]
			migrations = append(migrations, Migration{
				Version: version,
				Up:      filepath.Join(path, file.Name()),
				Down:    filepath.Join(path, version+".down.sql"),
			})
		}
	}
	return migrations, nil
}

func RunMigrations(provider *conn.PostgresProvider, migrationsPath string) error {
	log := logrus.New()
	db := provider.DB()

	// Создание таблицы schema_migrations
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		version VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`); err != nil {
		return fmt.Errorf("ошибка создания таблицы schema_migrations: %w", err)
	}

	migrations, err := getMigrations(migrationsPath)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		// Проверяем, была ли уже применена эта миграция
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", migration.Version).Scan(&count)
		if err != nil {
			return fmt.Errorf("ошибка проверки миграции %s: %w", migration.Version, err)
		}

		if count == 0 {
			// Читаем SQL из файла миграции
			content, err := os.ReadFile(migration.Up)
			if err != nil {
				return fmt.Errorf("не удалось прочитать файл миграции %s: %w", migration.Up, err)
			}

			// Выполняем миграцию
			log.Infof("Выполняется миграция: %s", migration.Version)
			if _, err := db.Exec(string(content)); err != nil {
				return fmt.Errorf("ошибка выполнения миграции %s: %w", migration.Version, err)
			}

			// Добавляем запись в schema_migrations
			if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				return fmt.Errorf("ошибка записи в schema_migrations: %w", err)
			}
		}
	}

	log.Info("Все миграции успешно применены.")
	return nil
}
