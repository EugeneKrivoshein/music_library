package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/EugeneKrivoshein/music_library/config"
	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/services"
	"github.com/gorilla/mux"
)

type Song struct {
	ID          int    `json:"id"`
	Group       string `json:"group"`
	Song        string `json:"song"`
	ReleaseDate string `json:"release_date,omitempty"`
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
	query := r.URL.Query()
	group := query.Get("group")
	song := query.Get("song")
	page, _ := strconv.Atoi(query.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	songs, err := h.SongService.GetSongs(group, song, page, limit)
	if err != nil {
		http.Error(w, "Ошибка получения песен: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
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
