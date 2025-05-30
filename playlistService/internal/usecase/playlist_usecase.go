package usecase

import (
	"context"
	"errors"
	domain2 "github.com/Zhan028/Music_Service/playlistService/internal/domain"
)

type PlaylistUseCase struct {
	repo domain2.PlaylistRepository
}

func NewPlaylistUseCase(repo domain2.PlaylistRepository) *PlaylistUseCase {
	return &PlaylistUseCase{
		repo: repo,
	}
}

func (uc *PlaylistUseCase) CreatePlaylist(ctx context.Context, name, userID, description string) (*domain2.Playlist, error) {
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
		Tracks:      []domain2.Track{},
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

	// Проверяем, что плейлист существует
	playlist, err := uc.repo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что трек с таким ID не существует в плейлисте
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
