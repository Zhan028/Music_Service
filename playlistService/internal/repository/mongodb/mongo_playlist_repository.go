package mongodb

import (
	"context"
	"errors"
	"fmt"
	domain2 "github.com/Zhan028/Music_Service/playlistService/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoPlaylistRepository struct {
	collection *mongo.Collection
}

// NewPlaylistRepository создает новый экземпляр репозитория для MongoDB
func NewPlaylistRepository(db *mongo.Database) domain2.PlaylistRepository {
	return &mongoPlaylistRepository{
		collection: db.Collection("playlists"),
	}
}

func (r *mongoPlaylistRepository) Create(ctx context.Context, playlist *domain2.Playlist) (*domain2.Playlist, error) {
	playlist, err := playlist.ToMongo()
	if err != nil {
		return nil, err
	}

	_, err = r.collection.InsertOne(ctx, playlist)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (r *mongoPlaylistRepository) GetByID(ctx context.Context, id string) (*domain2.Playlist, error) {
	var playlist domain2.Playlist

	// Сначала пробуем найти по строковому ID (если _id хранится как строка)
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&playlist)
	if err == nil {
		return playlist.FromMongo(), nil
	}

	// Если не найдено по строке, пробуем конвертировать в ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid playlist ID format")
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&playlist)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("playlist not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return playlist.FromMongo(), nil
}

func (r *mongoPlaylistRepository) GetByUserID(ctx context.Context, userID string) ([]*domain2.Playlist, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var playlists []*domain2.Playlist
	for cursor.Next(ctx) {
		var playlist domain2.Playlist
		if err := cursor.Decode(&playlist); err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist.FromMongo())
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func (r *mongoPlaylistRepository) Update(ctx context.Context, playlist *domain2.Playlist) (*domain2.Playlist, error) {
	objID, err := primitive.ObjectIDFromHex(playlist.ID)
	if err != nil {
		return nil, err
	}

	playlist.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        playlist.Name,
			"description": playlist.Description,
			"tracks":      playlist.Tracks,
			"updated_at":  playlist.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (r *mongoPlaylistRepository) AddTrack(ctx context.Context, playlistID string, track domain2.Track) (*domain2.Playlist, error) {
	objID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return nil, err
	}

	if track.ID == "" {
		track.ID = primitive.NewObjectID().Hex()
	}

	update := bson.M{
		"$push": bson.M{"tracks": track},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, playlistID)
}

func (r *mongoPlaylistRepository) RemoveTrack(ctx context.Context, playlistID string, trackID string) (*domain2.Playlist, error) {
	objID, err := primitive.ObjectIDFromHex(playlistID)
	if err != nil {
		return nil, err
	}

	update := bson.M{
		"$pull": bson.M{"tracks": bson.M{"_id": trackID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, playlistID)
}

func (r *mongoPlaylistRepository) Delete(ctx context.Context, id string, userID string) error {
	// Сначала получаем плейлист
	playlist, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Проверяем права доступа
	if playlist.UserID != userID {
		return errors.New("you don't have permission to delete this playlist")
	}

	// Пробуем удалить по строковому ID
	delResult, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}

	// Если не удалено, пробуем ObjectID
	if delResult.DeletedCount == 0 {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return errors.New("invalid playlist ID format")
		}

		delResult, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			return fmt.Errorf("database error: %v", err)
		}
		if delResult.DeletedCount == 0 {
			return errors.New("playlist not found")
		}
	}

	return nil
}
func (r *mongoPlaylistRepository) GetByName(ctx context.Context, name string) (*domain2.Playlist, error) {
	filter := bson.M{"name": name}

	var result domain2.Playlist
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // плейлист не найден — это не ошибка
		}
		return nil, err
	}

	return &result, nil
}
