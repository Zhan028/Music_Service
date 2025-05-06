package domain

import (
	"time"
)

// User represents the user entity
type User struct {
	ID        string     `bson:"_id,omitempty" json:"id"`
	Name      string     `bson:"name" json:"name"`
	Email     string     `bson:"email" json:"email"`
	Password  string     `bson:"password" json:"-"` // Password is not exposed in JSON responses
	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
}
