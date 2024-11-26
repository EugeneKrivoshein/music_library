package main

import (
	"net/http"

	"github.com/EugeneKrivoshein/music_library/config"
	"github.com/EugeneKrivoshein/music_library/internal/api"
	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/db/migrations"
	"github.com/EugeneKrivoshein/music_library/internal/handlers"
	"github.com/EugeneKrivoshein/music_library/internal/services"
	"github.com/sirupsen/logrus"
	_ "github.com/swaggo/swag/gen"
)

// @title Music Library API
// @version 1.0
// @description API для управления библиотекой песен
// @host localhost:8080
// @BasePath /

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	log.Info("Запуск приложения")
	log.Debug("Загрузка конфигурации")

	cfg, err := config.LoadConfig("config.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Infof("Конфигурация загружена: %+v", cfg)

	log.Debug("Подключение к базе данных")
	connect, err := conn.NewPostgresProvider(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer func() {
		log.Debug("Закрытие подключения к базе данных")
		connect.Close()
	}()
	log.Info("Подключение к базе данных успешно установлено")

	log.Debug("Запуск миграций базы данных")
	if err := migrations.RunMigrations(connect); err != nil {
		log.Fatalf("Ошибка выполнения миграций: %v", err)
	}
	log.Info("Миграции успешно выполнены")

	songService := services.NewSongService(connect, cfg)

	songHandler := handlers.NewSongHandler(connect, songService, cfg)

	// Создаем маршруты для API
	router := api.NewRouter(songHandler, connect)

	log.Infof("Сервер запущен на порту %s", cfg.ServerAddress)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
