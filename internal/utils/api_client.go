package utils

//для работы с внешним api
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/EugeneKrivoshein/music_library/config"
)

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// FetchSongDetails делает запрос к внешнему API и возвращает информацию о песне.
func FetchSongDetails(config *config.Config, group, song string) (*SongDetail, error) {
	apiURL := config.APIURL
	params := url.Values{}
	params.Add("group", group)
	params.Add("song", song)

	fullURL := fmt.Sprintf("%s/info?%s", apiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("внешний API вернул ошибку: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var songDetail SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа API: %w", err)
	}

	return &songDetail, nil
}
