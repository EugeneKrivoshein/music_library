package api

import (
	"net/http"

	"github.com/EugeneKrivoshein/music_library/internal/db/conn"
	"github.com/EugeneKrivoshein/music_library/internal/handlers"
	"github.com/gorilla/mux"
)

func NewRouter(songHandler *handlers.SongHandler, dbProvider *conn.PostgresProvider) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Сервер работает!"))
	}).Methods("GET")
	router.HandleFunc("/songs", songHandler.GetSongs).Methods("GET")
	router.HandleFunc("/songs/{id:[0-9]+}", songHandler.GetSongText).Methods("GET")
	router.HandleFunc("/songs/add", songHandler.AddSongWithAPI).Methods("POST")
	router.HandleFunc("/songs/{id:[0-9]+}", songHandler.UpdateSong).Methods("PUT")
	router.HandleFunc("/songs/{id:[0-9]+}", songHandler.DeleteSong).Methods("DELETE")

	return router
}
