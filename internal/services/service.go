package services

import (
	"database/sql"
	"fmt"

	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/utils"
	"github.com/sirupsen/logrus"
)

type SongService struct {
	dbProvider *conn.PostgresProvider
}

func NewSongService(provider *conn.PostgresProvider) *SongService {
	return &SongService{dbProvider: provider}
}

var log = logrus.New()

func (s *SongService) GetSongs(group, song string, page, limit int) ([]map[string]interface{}, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, group_name, song_name, release_date
		FROM songs
		WHERE ($1 = '' OR group_name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR song_name ILIKE '%' || $2 || '%')
		ORDER BY id LIMIT $3 OFFSET $4`

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
		var group, song string
		var releaseDate sql.NullString
		if err := rows.Scan(&id, &group, &song, &releaseDate); err != nil {
			log.Errorf("Ошибка сканирования строки: %v", err)
			return nil, err
		}

		songData := map[string]interface{}{
			"id":           id,
			"group":        group,
			"song":         song,
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
	query := `
		UPDATE songs
		SET group_name = COALESCE(NULLIF($1, ''), group_name),
		    song_name = COALESCE(NULLIF($2, ''), song_name),
		    release_date = COALESCE($3::DATE, release_date),
		    text = COALESCE($4, text),
		    link = COALESCE($5, link),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6`

	db := s.dbProvider.DB()
	_, err := db.Exec(query, group, song, releaseDate, text, link, id)
	if err != nil {
		log.Errorf("Ошибка обновления песни с ID %d: %v", id, err)
		return fmt.Errorf("ошибка обновления песни: %w", err)
	}
	log.Infof("Песня с ID %d успешно обновлена", id)
	return nil
}

func (s *SongService) AddSongWithAPI(group, song string) error {
	apiURL := "https://api.com"
	details, err := utils.FetchSongDetails(apiURL, group, song)
	if err != nil {
		log.Errorf("Ошибка вызова внешнего API: %v", err)
		return fmt.Errorf("ошибка вызова внешнего API: %w", err)
	}

	query := `
		INSERT INTO songs (group_name, song_name, release_date, lyrics, link)
		VALUES ($1, $2, $3, $4, $5)`

	db := s.dbProvider.DB()
	_, err = db.Exec(query, group, song, details.ReleaseDate, details.Text, details.Link)
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
