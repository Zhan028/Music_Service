package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Track struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Artist    string             `bson:"artist"`
	Album     string             `bson:"album"`
	Duration  int32              `bson:"duration_sec"`
	CreatedAt int64              `bson:"created_at"`
}
