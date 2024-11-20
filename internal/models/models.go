package models

// Song represents a song in the library.
// @Description Представляет песню в библиотеке
type Song struct {
	ID          int    `db:"id"`
	GroupName   string `db:"group_name"`
	SongName    string `db:"song_name"`
	ReleaseDate string `db:"release_date"`
	Text        string `db:"text"`
	Link        string `db:"link"`
}
