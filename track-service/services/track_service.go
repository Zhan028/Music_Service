package services

import (
	"context"
	localKafka "github.com/Zhan028/Music_Service/track-service/kafka"
	"github.com/Zhan028/Music_Service/track-service/models"
	pb "github.com/Zhan028/Music_Service/track-service/proto"

	"github.com/Zhan028/Music_Service/track-service/repositories"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrackGRPCService struct {
	repo *repositories.TrackRepo
	pb.UnimplementedTrackServiceServer
}

func NewTrackGRPCService(repo *repositories.TrackRepo) *TrackGRPCService {
	return &TrackGRPCService{repo: repo}
}

// Kafka writer (можно вынести в init или конструктор)
var kafkaWriter = &kafka.Writer{
	Addr:     kafka.TCP("localhost:9092"),
	Balancer: &kafka.LeastBytes{},
}

func (s *TrackGRPCService) CreateTrack(ctx context.Context, req *pb.CreateTrackRequest) (*pb.CreateTrackResponse, error) {
	track := models.Track{
		ID:       primitive.NewObjectID(),
		Title:    req.GetTitle(),
		Artist:   req.GetArtist(),
		Album:    req.GetAlbum(),
		Duration: req.GetDurationSec(),
	}

	// CreatedAt внутри репозитория
	if _, err := s.repo.CreateTrack(ctx, track); err != nil {
		return nil, err
	}

	// Отправляем событие в Kafka
	localKafka.PublishMessage(ctx, kafkaWriter, "track.created", track.ID.Hex(), track)

	return &pb.CreateTrackResponse{
		Track: toProto(track),
	}, nil
}

func (s *TrackGRPCService) GetTrackByID(ctx context.Context, req *pb.GetTrackByIDRequest) (*pb.GetTrackByIDResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, err
	}

	track, err := s.repo.GetTrackByID(ctx, objID)
	if err != nil {
		return nil, err
	}

	return &pb.GetTrackByIDResponse{
		Track: toProto(track),
	}, nil
}

func (s *TrackGRPCService) GetAllTracks(ctx context.Context, req *pb.GetAllTracksRequest) (*pb.GetAllTracksResponse, error) {
	filter := bson.M{}
	if title := req.GetTitle(); title != "" {
		filter["title"] = bson.M{"$regex": title, "$options": "i"}
	}
	if artist := req.GetArtist(); artist != "" {
		filter["artist"] = bson.M{"$regex": artist, "$options": "i"}
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 10
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	skip := (page - 1) * limit

	tracks, err := s.repo.GetAllTracks(ctx, filter, limit, skip)
	if err != nil {
		return nil, err
	}

	protoTracks := make([]*pb.Track, len(tracks))
	for i, t := range tracks {
		protoTracks[i] = toProto(t)
	}

	return &pb.GetAllTracksResponse{Tracks: protoTracks}, nil
}

func (s *TrackGRPCService) UpdateTrack(ctx context.Context, req *pb.UpdateTrackRequest) (*pb.UpdateTrackResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, err
	}

	updateData := bson.M{}
	if title := req.GetTitle(); title != "" {
		updateData["title"] = title
	}
	if artist := req.GetArtist(); artist != "" {
		updateData["artist"] = artist
	}
	if album := req.GetAlbum(); album != "" {
		updateData["album"] = album
	}
	if dur := req.GetDurationSec(); dur != 0 {
		updateData["duration_sec"] = dur
	}

	if len(updateData) == 0 {
		return &pb.UpdateTrackResponse{Message: "No fields to update"}, nil
	}

	if err := s.repo.UpdateTrack(ctx, objID, updateData); err != nil {
		return nil, err
	}

	return &pb.UpdateTrackResponse{Message: "Track updated successfully"}, nil
}

func (s *TrackGRPCService) DeleteTrack(ctx context.Context, req *pb.DeleteTrackRequest) (*pb.DeleteTrackResponse, error) {
	objID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteTrack(ctx, objID); err != nil {
		return nil, err
	}

	return &pb.DeleteTrackResponse{Message: "Track deleted successfully"}, nil
}

func toProto(t models.Track) *pb.Track {
	return &pb.Track{
		Id:          t.ID.Hex(),
		Title:       t.Title,
		Artist:      t.Artist,
		Album:       t.Album,
		DurationSec: t.Duration,
		CreatedAt:   t.CreatedAt,
	}
}
