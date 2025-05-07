package grpc

import (
	"context"
	"github.com/Zhan028/Music_Service/internal/domain"
	"github.com/Zhan028/Music_Service/internal/usecase"
	"github.com/Zhan028/Music_Service/proto"

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

func (s *PlaylistServer) CreatePlaylist(ctx context.Context, req *proto.CreatePlaylistRequest) (*proto.Playlist, error) {
	playlist, err := s.useCase.CreatePlaylist(ctx, req.Name, req.UserId, req.Description)
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
	err := s.useCase.DeletePlaylist(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete playlist: %v", err)
	}

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
