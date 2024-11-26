package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EugeneKrivoshein/music_library/config"
	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/services"
	"github.com/gorilla/mux"
)

type Song struct {
	ID          int    `json:"id"`
	GroupName   string `json:"group_name"`
	SongName    string `json:"song_name"`
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type SongHandler struct {
	SongService *services.SongService
	dbProvider  *conn.PostgresProvider
	Config      *config.Config
}

func NewSongHandler(provider *conn.PostgresProvider, service *services.SongService, cfg *config.Config) *SongHandler {
	return &SongHandler{
		dbProvider:  provider,
		SongService: service,
		Config:      cfg,
	}
}

// GetSongs godoc
// @Summary Получить список песен
// @Description Возвращает список песен с фильтрацией по группе и названию песни, а также поддержкой пагинации.
// @Tags Songs
// @Accept json
// @Produce json
// @Param group query string false "Название группы" default()
// @Param song query string false "Название песни" default()
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {array} handlers.Song "Список песен"
// @Failure 400 {string} string "Некорректные параметры запроса"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /songs [get]
func (h *SongHandler) GetSongs(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметры из запроса
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Преобразуем параметры страницы и лимита
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1 // если страница некорректна, то начинаем с первой
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10 // если лимит некорректен, то устанавливаем значение по умолчанию
	}

	// Получаем список песен через сервис
	songs, err := h.SongService.GetSongs(group, song, page, limit)
	if err != nil {
		http.Error(w, "Ошибка получения песен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразуем данные в нужный формат
	var response []map[string]interface{}
	for _, songData := range songs {
		// Извлекаем нужные поля из карты
		text, ok := songData["text"].(string) // здесь мы предполагаем, что поле text будет в карте
		if !ok {
			text = "" // если текст не найден, оставляем пустым
		}

		// Разделяем текст песни на куплеты, если текст существует
		verses := []string{}
		if text != "" {
			verses = strings.Split(text, "\n")
		}

		// Пагинация по куплетам: определяем куплеты для текущей страницы
		startIndex := (page - 1) * limit
		endIndex := startIndex + limit

		// Проверяем, что индексы не выходят за пределы массива куплетов
		if startIndex > len(verses) {
			verses = []string{} // Если страницы не существует, возвращаем пустой список
		} else if endIndex > len(verses) {
			verses = verses[startIndex:] // Если конец выходит за пределы, берем только оставшиеся куплеты
		} else {
			verses = verses[startIndex:endIndex]
		}

		// Обновляем текст песни для текущей страницы (по куплетам)
		songData["text"] = strings.Join(verses, "\n")

		// Добавляем песню в ответ
		response = append(response, songData)
	}

	// Отправляем JSON-ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка кодирования ответа: "+err.Error(), http.StatusInternalServerError)
	}
}

// GetSongText godoc
// @Summary Получить текст песни
// @Description Возвращает текст песни построчно.
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Success 200 {string} string "Текст песни"
// @Failure 400 {string} string "Некорректный ID"
// @Failure 500 {string} string "Ошибка получения текста песни"
// @Router /songs/{id} [get]
func (h *SongHandler) GetSongText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	text, err := h.SongService.GetSongText(id)
	if err != nil {
		http.Error(w, "Ошибка получения текста песни: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(text)
}

// AddSongWithAPI добавляет песню через внешнее API.
// @Summary Добавить песню через API
// @Description Добавляет новую песню, используя данные внешнего API.
// @Tags Songs
// @Accept json
// @Produce json
// @Param input body Song true "Данные песни"
// @Success 201 {string} string "Песня успешно добавлена"
// @Failure 400 {string} string "Некорректные входные данные"
// @Failure 500 {string} string "Ошибка добавления песни"
// @Router /songs/add [post]
func (h *SongHandler) AddSongWithAPI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Group string `json:"group"`
		Song  string `json:"song"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Group == "" || input.Song == "" {
		http.Error(w, "Некорректные входные данные", http.StatusBadRequest)
		return
	}

	if err := h.SongService.AddSongWithAPI(h.Config, input.Group, input.Song); err != nil {
		http.Error(w, "Ошибка добавления песни: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Песня успешно добавлена"))
}

// UpdateSong обновляет данные песни.
// @Summary Обновить песню
// @Description Обновляет данные песни по её ID.
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param input body Song true "Обновляемые данные песни"
// @Success 200 {string} string "Песня успешно обновлена"
// @Failure 400 {string} string "Некорректный ID или формат данных"
// @Failure 500 {string} string "Ошибка обновления песни"
// @Router /songs/{id} [put]
func (h *SongHandler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Group       string  `json:"group,omitempty"`
		Song        string  `json:"song,omitempty"`
		ReleaseDate *string `json:"release_date,omitempty"` // Формат даты: "YYYY-MM-DD"
		Text        *string `json:"text,omitempty"`
		Link        *string `json:"link,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Некорректный формат данных", http.StatusBadRequest)
		return
	}

	// Валидация полей, которые не могут быть пустыми
	if input.Group == "" {
		http.Error(w, "Группа не может быть пустой", http.StatusBadRequest)
		return
	}
	if input.Song == "" {
		http.Error(w, "Название песни не может быть пустым", http.StatusBadRequest)
		return
	}

	// Проверяем правильность формата даты, если она передана
	if input.ReleaseDate != nil {
		if _, err := time.Parse("2006-01-02", *input.ReleaseDate); err != nil {
			http.Error(w, "Некорректный формат даты", http.StatusBadRequest)
			return
		}
	}

	if err := h.SongService.UpdateSong(id, input.Group, input.Song, input.ReleaseDate, input.Text, input.Link); err != nil {
		http.Error(w, "Ошибка обновления песни: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Песня успешно обновлена"))
}

// DeleteSong godoc
// @Summary Удалить песню
// @Description Удаляет песню из библиотеки по ID.
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Success 204 {string} string "Песня успешно удалена"
// @Failure 400 {string} string "Некорректный ID"
// @Failure 500 {string} string "Ошибка удаления песни"
// @Router /songs/{id} [delete]
func (h *SongHandler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	if err := h.SongService.DeleteSong(id); err != nil {
		http.Error(w, "Ошибка удаления песни: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
