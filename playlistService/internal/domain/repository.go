package domain

import "context"

// PlaylistRepository описывает методы для работы с хранилищем плейлистов
type PlaylistRepository interface {
	// Create создает новый плейлист
	Create(ctx context.Context, playlist *Playlist) (*Playlist, error)

	// GetByID получает плейлист по ID
	GetByID(ctx context.Context, id string) (*Playlist, error)

	// GetByUserID получает все плейлисты пользователя
	GetByUserID(ctx context.Context, userID string) ([]*Playlist, error)

	// Update обновляет существующий плейлист
	Update(ctx context.Context, playlist *Playlist) (*Playlist, error)

	// AddTrack добавляет трек в плейлист
	AddTrack(ctx context.Context, playlistID string, track Track) (*Playlist, error)

	// RemoveTrack удаляет трек из плейлиста
	RemoveTrack(ctx context.Context, playlistID string, trackID string) (*Playlist, error)

	// Delete удаляет плейлист
	Delete(ctx context.Context, id string, userID string) error

	GetByName(ctx context.Context, name string) (*Playlist, error)
}
