package grpc

import (
	"context"
	"github.com/Zhan028/Music_Service/playlistService/internal/domain"
	"github.com/Zhan028/Music_Service/playlistService/internal/usecase"
	"github.com/Zhan028/Music_Service/playlistService/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PlaylistServer struct {
	proto.UnimplementedPlaylistServiceServer
	useCase *usecase.PlaylistUseCase
}

func NewPlaylistServer(useCase *usecase.PlaylistUseCase) *PlaylistServer {
	return &PlaylistServer{
		useCase: useCase,
	}
}

func convertProtoTracksToDomainTracks(protoTracks []*proto.Track) []*domain.Track {
	var domainTracks []*domain.Track

	for _, track := range protoTracks {
		domainTracks = append(domainTracks, &domain.Track{
			ID:       primitive.NewObjectID().Hex(), // генерируем ID
			Title:    track.Title,
			Artist:   track.Artist,
			Duration: track.Duration,
			Album:    track.Album,
		})
	}
	return domainTracks
}

func (s *PlaylistServer) CreatePlaylist(ctx context.Context, req *proto.CreatePlaylistRequest) (*proto.Playlist, error) {
	// Логируем запрос, чтобы увидеть, что мы получаем
	log.Printf("Received CreatePlaylist request: %+v\n", req)

	// Конвертируем proto треки в доменные
	domainTracks := convertProtoTracksToDomainTracks(req.Tracks)

	// Логируем конвертированные треки
	log.Printf("Converted domain tracks: %+v\n", domainTracks)

	// Создаем плейлист с конвертированными треками
	playlist, err := s.useCase.CreatePlaylist(ctx, req.Name, req.UserId, req.Description, domainTracks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create playlist: %v", err)
	}

	return convertDomainToProto(playlist), nil
}

func (s *PlaylistServer) GetPlaylist(ctx context.Context, req *proto.GetPlaylistRequest) (*proto.Playlist, error) {
	playlist, err := s.useCase.GetPlaylist(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "playlist not found: %v", err)
	}

	return convertDomainToProto(playlist), nil
}

func (s *PlaylistServer) GetUserPlaylists(ctx context.Context, req *proto.GetUserPlaylistsRequest) (*proto.PlaylistList, error) {
	playlists, err := s.useCase.GetUserPlaylists(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user playlists: %v", err)
	}

	protoPlaylists := &proto.PlaylistList{
		Playlists: make([]*proto.Playlist, 0, len(playlists)),
	}

	for _, playlist := range playlists {
		protoPlaylists.Playlists = append(protoPlaylists.Playlists, convertDomainToProto(playlist))
	}

	return protoPlaylists, nil
}

func (s *PlaylistServer) AddTrackToPlaylist(ctx context.Context, req *proto.AddTrackRequest) (*proto.Playlist, error) {
	track := domain.Track{
		ID:       req.Track.Id,
		Title:    req.Track.Title,
		Artist:   req.Track.Artist,
		Duration: req.Track.Duration,
		Album:    req.Track.Album,
	}

	playlist, err := s.useCase.AddTrackToPlaylist(ctx, req.PlaylistId, track)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add track to playlist: %v", err)
	}

	return convertDomainToProto(playlist), nil
}

func (s *PlaylistServer) RemoveTrackFromPlaylist(ctx context.Context, req *proto.RemoveTrackRequest) (*proto.Playlist, error) {
	playlist, err := s.useCase.RemoveTrackFromPlaylist(ctx, req.PlaylistId, req.TrackId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove track from playlist: %v", err)
	}

	return convertDomainToProto(playlist), nil
}

func (s *PlaylistServer) DeletePlaylist(ctx context.Context, req *proto.DeletePlaylistRequest) (*proto.DeletePlaylistResponse, error) {
	// Добавляем логирование для отладки
	log.Printf("Delete request for playlist ID: %s, user ID: %s", req.Id, req.UserId)

	// 1. Проверяем существование плейлиста
	playlist, err := s.useCase.GetPlaylist(ctx, req.Id)
	if err != nil {
		log.Printf("Playlist not found error: %v", err)
		return nil, status.Errorf(codes.NotFound, "playlist not found")
	}

	// 2. Проверяем права доступа
	if playlist.UserID != req.UserId {
		log.Printf("Permission denied: playlist owned by %s, requested by %s",
			playlist.UserID, req.UserId)
		return nil, status.Errorf(codes.PermissionDenied, "you don't own this playlist")
	}

	// 3. Удаляем плейлист
	if err := s.useCase.DeletePlaylist(ctx, req.Id, req.UserId); err != nil {
		log.Printf("Delete error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete playlist")
	}

	log.Printf("Playlist %s deleted successfully", req.Id)
	return &proto.DeletePlaylistResponse{Success: true}, nil
}

// Вспомогательные функции для конвертации между моделями
func convertDomainToProto(playlist *domain.Playlist) *proto.Playlist {
	protoTracks := make([]*proto.Track, 0, len(playlist.Tracks))

	for _, track := range playlist.Tracks {
		protoTracks = append(protoTracks, &proto.Track{
			Id:       track.ID,
			Title:    track.Title,
			Artist:   track.Artist,
			Duration: track.Duration,
			Album:    track.Album,
		})
	}

	return &proto.Playlist{
		Id:          playlist.ID,
		Name:        playlist.Name,
		UserId:      playlist.UserID,
		Description: playlist.Description,
		Tracks:      protoTracks,
		CreatedAt:   playlist.CreatedAt.Unix(),
		UpdatedAt:   playlist.UpdatedAt.Unix(),
	}
}
