package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Track представляет музыкальный трек
type Track struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Title    string `json:"title" bson:"title"`
	Artist   string `json:"artist" bson:"artist"`
	Duration int32  `json:"duration" bson:"duration"` // в секундах
	Album    string `json:"album" bson:"album"`
}

// Playlist представляет плейлист
type Playlist struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Name        string    `json:"name" bson:"name"`
	UserID      string    `json:"user_id" bson:"user_id"`
	Description string    `json:"description" bson:"description"`
	Tracks      []Track   `json:"tracks" bson:"tracks"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// ToMongo конвертирует ID в ObjectID для MongoDB
func (p *Playlist) ToMongo() (*Playlist, error) {
	if p.ID == "" {
		p.ID = primitive.NewObjectID().Hex()
	}

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()

	return p, nil
}

// FromMongo восстанавливает поля после получения из MongoDB
func (p *Playlist) FromMongo() *Playlist {
	return p
}
