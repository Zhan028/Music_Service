package usecase

import (
	"context"
	"encoding/json"
	"errors"
	domain2 "github.com/Zhan028/Music_Service/playlistService/internal/domain"
	"github.com/segmentio/kafka-go"
	"log"
)

type PlaylistUseCase struct {
	repo domain2.PlaylistRepository
}

func NewPlaylistUseCase(repo domain2.PlaylistRepository) *PlaylistUseCase {
	return &PlaylistUseCase{
		repo: repo,
	}
}

func (uc *PlaylistUseCase) CreatePlaylist(ctx context.Context, name, userID, description string, tracks []*domain2.Track) (*domain2.Playlist, error) {
	if name == "" {
		return nil, errors.New("playlist name cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	playlist := &domain2.Playlist{
		Name:        name,
		UserID:      userID,
		Description: description,
		Tracks:      tracks,
	}

	return uc.repo.Create(ctx, playlist)
}

func (uc *PlaylistUseCase) GetPlaylist(ctx context.Context, id string) (*domain2.Playlist, error) {
	if id == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *PlaylistUseCase) GetUserPlaylists(ctx context.Context, userID string) ([]*domain2.Playlist, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	return uc.repo.GetByUserID(ctx, userID)
}

func (uc *PlaylistUseCase) AddTrackToPlaylist(ctx context.Context, playlistID string, track domain2.Track) (*domain2.Playlist, error) {
	if playlistID == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	if track.Title == "" || track.Artist == "" {
		return nil, errors.New("track title and artist cannot be empty")
	}

	playlist, err := uc.repo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	for _, t := range playlist.Tracks {
		if t.ID == track.ID && track.ID != "" {
			return nil, errors.New("track already exists in playlist")
		}
	}

	return uc.repo.AddTrack(ctx, playlistID, track)
}

func (uc *PlaylistUseCase) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) (*domain2.Playlist, error) {
	if playlistID == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	if trackID == "" {
		return nil, errors.New("track ID cannot be empty")
	}

	return uc.repo.RemoveTrack(ctx, playlistID, trackID)
}

func (uc *PlaylistUseCase) DeletePlaylist(ctx context.Context, id, userID string) error {
	if id == "" {
		return errors.New("playlist ID cannot be empty")
	}

	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	return uc.repo.Delete(ctx, id, userID)
}

// UpdatePlaylist обновляет информацию о плейлисте
func (uc *PlaylistUseCase) UpdatePlaylist(ctx context.Context, id, userID, name, description string, isPublic bool) (*domain2.Playlist, error) {
	if id == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	// Проверяем существование плейлиста и права доступа
	playlist, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if playlist.UserID != userID {
		return nil, errors.New("you don't have permission to update this playlist")
	}

	// Создаем обновленную версию плейлиста
	updatedPlaylist := &domain2.Playlist{
		ID:          playlist.ID,
		Name:        name,
		UserID:      userID,
		Description: description,
		Tracks:      playlist.Tracks,

		CreatedAt: playlist.CreatedAt,
	}

	return uc.repo.Update(ctx, updatedPlaylist)
}

func (uc *PlaylistUseCase) AddToNewPlaylist(ctx context.Context, message kafka.Message) error {
	var track domain2.Track
	if err := json.Unmarshal(message.Value, &track); err != nil {
		log.Printf("unmarshal error: %v", err)
		return err
	}
	const playlistName = "Новинки"

	playlist, err := uc.repo.GetByName(ctx, playlistName)
	if err != nil {
		return err
	}

	if playlist == nil {
		newPlaylist := &domain2.Playlist{
			Name:   playlistName,
			UserID: "system", // специальный ID для системных плейлистов
			Tracks: []*domain2.Track{&track},
		}

		_, err := uc.repo.Create(ctx, newPlaylist)
		return err
	}

	// Проверяем, нет ли уже такого трека в плейлисте
	for _, t := range playlist.Tracks {
		if t.ID == track.ID && track.ID != "" {
			return nil // трек уже есть, ничего не делаем
		}
	}

	_, err = uc.repo.AddTrack(ctx, playlist.ID, track)
	return err
}
