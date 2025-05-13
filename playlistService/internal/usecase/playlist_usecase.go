package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domain2 "github.com/Zhan028/Music_Service/playlistService/internal/domain"
	"github.com/Zhan028/Music_Service/playlistService/internal/redis"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

type PlaylistUseCase struct {
	repo  domain2.PlaylistRepository
	redis redis.Redis
}

func NewPlaylistUseCase(repo domain2.PlaylistRepository, redis redis.Redis) *PlaylistUseCase {
	return &PlaylistUseCase{
		repo:  repo,
		redis: redis,
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

	result, err := uc.repo.Create(ctx, playlist)

	if err == nil {
		uc.redis.Delete("user_playlists:" + userID)
	}

	return result, err
}

func (uc *PlaylistUseCase) GetPlaylist(ctx context.Context, id string) (*domain2.Playlist, error) {
	if id == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *PlaylistUseCase) GetUserPlaylists(ctx context.Context, userID string) ([]*domain2.Playlist, error) {
	log.Println(" GetUserPlaylists called")
	start := time.Now()

	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	cacheKey := "user_playlists:" + userID

	cached, err := uc.redis.Get(cacheKey)
	if err == nil {
		log.Println(" Redis HIT")
		var playlists []*domain2.Playlist
		if err := json.Unmarshal(cached, &playlists); err == nil {
			log.Printf(" Duration (HIT): %v", time.Since(start))
			return playlists, nil
		}
	}

	log.Println("Redis MISS — получаю из базы")
	playlists, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	//cache
	err = uc.redis.Set(cacheKey, playlists, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	log.Printf(" Duration (MISS): %v", time.Since(start))
	return playlists, nil
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

	// Удаляем кэш по userID владельца плейлиста
	uc.redis.Delete("user_playlists:" + playlist.UserID)

	// Добавляем трек
	return uc.repo.AddTrack(ctx, playlist.ID, track)
}

func (uc *PlaylistUseCase) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) (*domain2.Playlist, error) {
	if playlistID == "" {
		return nil, errors.New("playlist ID cannot be empty")
	}

	if trackID == "" {
		return nil, errors.New("track ID cannot be empty")
	}

	// Получаем плейлист, чтобы узнать userID
	playlist, err := uc.repo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	// Удаляем кэш по userID
	uc.redis.Delete("user_playlists:" + playlist.UserID)

	// Удаляем трек
	return uc.repo.RemoveTrack(ctx, playlistID, trackID)
}

func (uc *PlaylistUseCase) DeletePlaylist(ctx context.Context, id, userID string) error {
	if id == "" {
		return errors.New("playlist ID cannot be empty")
	}
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if err := uc.repo.Delete(ctx, id, userID); err != nil {
		return err
	}

	uc.redis.Delete("user_playlists:" + userID)
	return nil
}
func (uc *PlaylistUseCase) AddToNewPlaylist(ctx context.Context, message kafka.Message) error {
	var track domain2.Track
	if err := json.Unmarshal(message.Value, &track); err != nil {
		log.Printf("unmarshal error: %v", err)
		return err
	}
	const playlistName = "Новинки-2"
	const userID = "system"

	playlist, _ := uc.repo.GetByName(ctx, playlistName)
	fmt.Println(playlist)

	if playlist == nil {
		newPlaylist := &domain2.Playlist{
			ID:     primitive.NewObjectID().Hex(),
			Name:   playlistName,
			UserID: userID,
			Tracks: []*domain2.Track{&track},
		}

		_, err := uc.repo.Create(ctx, newPlaylist)
		if err == nil {
			uc.redis.Delete("user_playlists:" + userID)
			log.Printf(" Redis cache invalidated: user_playlists:%s", userID)
		}
		return err
	}

	_, err := uc.repo.AddTrack(ctx, playlist.ID, track)
	if err == nil {
		uc.redis.Delete("user_playlists:" + playlist.UserID)
		log.Printf("Redis cache invalidated: user_playlists:%s", playlist.UserID)
	}
	return err
}
