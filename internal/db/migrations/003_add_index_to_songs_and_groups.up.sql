-- Индексы для таблицы songs
CREATE INDEX idx_songs_group_id ON songs (group_id);
CREATE INDEX idx_songs_song_name ON songs (song_name);
CREATE INDEX idx_songs_release_date ON songs (release_date);
-- Индексы для таблицы groups
CREATE INDEX idx_groups_name ON groups (name);