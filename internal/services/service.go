package services

import (
	"database/sql"
	"fmt"

	"github.com/EugeneKrivoshein/music_library/config"
	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/utils"
	"github.com/sirupsen/logrus"
)

type SongService struct {
	dbProvider *conn.PostgresProvider
	APIURL     string
}

func NewSongService(provider *conn.PostgresProvider, config *config.Config) *SongService {
	return &SongService{
		dbProvider: provider,
		APIURL:     config.APIURL,
	}
}

var log = logrus.New()

func (s *SongService) GetSongs(group, song string, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit
	query := `
		SELECT s.id, g.group_name, s.song_name, s.release_date
		FROM songs s
		JOIN groups g ON s.group_id = g.id
		WHERE ($1 = '' OR g.group_name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR s.song_name ILIKE '%' || $2 || '%')
		ORDER BY s.id LIMIT $3 OFFSET $4`

	db := s.dbProvider.DB()
	rows, err := db.Query(query, group, song, limit, offset)
	if err != nil {
		log.Errorf("Ошибка выполнения запроса: %v", err)
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	songs := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var groupName, songName string
		var releaseDate sql.NullString
		if err := rows.Scan(&id, &groupName, &songName, &releaseDate); err != nil {
			log.Errorf("Ошибка сканирования строки: %v", err)
			return nil, err
		}

		songData := map[string]interface{}{
			"id":           id,
			"group":        groupName,
			"song":         songName,
			"release_date": releaseDate.String,
		}
		songs = append(songs, songData)
	}
	log.Infof("Найдено %d песен", len(songs))
	return songs, nil
}

func (s *SongService) GetSongText(id int) (string, error) {
	query := `SELECT text FROM songs WHERE id = $1`
	var text string
	db := s.dbProvider.DB()
	err := db.QueryRow(query, id).Scan(&text)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warnf("Песня с ID %d не найдена", id)
			return "", fmt.Errorf("песня с id %d не найдена", id)
		}
		log.Errorf("Ошибка получения текста песни: %v", err)
		return "", fmt.Errorf("ошибка получения текста песни: %w", err)
	}
	log.Infof("Текст песни с ID %d успешно получен", id)
	return text, nil
}

func (s *SongService) UpdateSong(id int, group, song string, releaseDate *string, text, link *string) error {
	db := s.dbProvider.DB()

	// Получаем group_id, если передано новое название группы
	var groupID *int
	if group != "" {
		groupID = new(int)
		err := db.QueryRow(`SELECT id FROM groups WHERE group_name = $1`, group).Scan(groupID)
		if err == sql.ErrNoRows {
			// Добавляем группу, если её нет
			err = db.QueryRow(`INSERT INTO groups (group_name) VALUES ($1) RETURNING id`, group).Scan(groupID)
			if err != nil {
				log.Errorf("Ошибка добавления группы: %v", err)
				return fmt.Errorf("ошибка добавления группы: %w", err)
			}
		} else if err != nil {
			log.Errorf("Ошибка проверки группы: %v", err)
			return fmt.Errorf("ошибка проверки группы: %w", err)
		}
	}

	// Обновление песни
	query := `
		UPDATE songs
		SET group_id = COALESCE($1, group_id),
		    song_name = COALESCE(NULLIF($2, ''), song_name),
		    release_date = COALESCE($3::DATE, release_date),
		    text = COALESCE($4, text),
		    link = COALESCE($5, link),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6`
	_, err := db.Exec(query, groupID, song, releaseDate, text, link, id)
	if err != nil {
		log.Errorf("Ошибка обновления песни с ID %d: %v", id, err)
		return fmt.Errorf("ошибка обновления песни: %w", err)
	}
	log.Infof("Песня с ID %d успешно обновлена", id)
	return nil
}

func (s *SongService) AddSongWithAPI(config *config.Config, group, song string) error {
	db := s.dbProvider.DB()

	// Проверка существования группы
	var groupID int
	err := db.QueryRow(`SELECT id FROM groups WHERE group_name = $1`, group).Scan(&groupID)
	if err == sql.ErrNoRows {
		// Группа не найдена, добавляем
		err = db.QueryRow(`INSERT INTO groups (group_name) VALUES ($1) RETURNING id`, group).Scan(&groupID)
		if err != nil {
			log.Errorf("Ошибка добавления группы: %v", err)
			return fmt.Errorf("ошибка добавления группы: %w", err)
		}
	} else if err != nil {
		log.Errorf("Ошибка проверки группы: %v", err)
		return fmt.Errorf("ошибка проверки группы: %w", err)
	}

	// Получение деталей песни из внешнего API
	details, err := utils.FetchSongDetails(config, group, song)
	if err != nil {
		log.Errorf("Ошибка вызова внешнего API: %v", err)
		return fmt.Errorf("ошибка вызова внешнего API: %w", err)
	}

	// Добавление песни
	query := `
		INSERT INTO songs (group_id, song_name, release_date, lyrics, link)
		VALUES ($1, $2, $3, $4, $5)`
	_, err = db.Exec(query, groupID, song, details.ReleaseDate, details.Text, details.Link)
	if err != nil {
		log.Errorf("Ошибка сохранения песни в базу: %v", err)
		return fmt.Errorf("ошибка сохранения песни: %w", err)
	}

	log.Infof("Песня %s - %s успешно добавлена", group, song)
	return nil
}

// Удаление песни по ID
func (s *SongService) DeleteSong(id int) error {
	query := `DELETE FROM songs WHERE id = $1`

	db := s.dbProvider.DB()
	_, err := db.Exec(query, id)
	if err != nil {
		log.Errorf("Ошибка удаления песни с ID %d: %v", id, err)
		return fmt.Errorf("ошибка удаления песни: %w", err)
	}
	log.Infof("Песня с ID %d успешно удалена", id)
	return nil
}
