package repositories

import (
	"context"
	"time"

	"github.com/Zhan028/Music_Service/track-service/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TrackRepo struct {
	collection *mongo.Collection
}

func NewTrackRepo(db *mongo.Database) *TrackRepo {
	return &TrackRepo{
		collection: db.Collection("tracks"),
	}
}

func (r *TrackRepo) CreateTrack(ctx context.Context, track models.Track) (*mongo.InsertOneResult, error) {
	track.CreatedAt = time.Now().Unix()
	return r.collection.InsertOne(ctx, track)
}

func (r *TrackRepo) GetTrackByID(ctx context.Context, id primitive.ObjectID) (models.Track, error) {
	var track models.Track
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&track)
	return track, err
}

func (r *TrackRepo) GetAllTracks(ctx context.Context, filter bson.M, limit int64, skip int64) ([]models.Track, error) {
	var tracks []models.Track

	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(limit)
	}
	if skip > 0 {
		findOptions.SetSkip(skip)
	}

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var track models.Track
		if err := cursor.Decode(&track); err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (r *TrackRepo) UpdateTrack(ctx context.Context, id primitive.ObjectID, updateData bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": updateData})
	return err
}

func (r *TrackRepo) DeleteTrack(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
