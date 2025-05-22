package usecase_test

import (
	"context"
	"github.com/Zhan028/Music_Service/userService/internal/repository"
	"github.com/Zhan028/Music_Service/userService/internal/usecase"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestMongo(t *testing.T) (*mongo.Client, *mongo.Database, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)

	db := client.Database("ap2")
	err = db.Collection("users").Drop(ctx)
	assert.NoError(t, err)

	return client, db, func() {
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
		cancel()
	}
}

func TestUserUseCase_GetByID(t *testing.T) {
	_, db, teardown := setupTestMongo(t)
	defer teardown()

	objID := primitive.NewObjectID()

	coll := db.Collection("users")
	_, err := coll.InsertOne(context.Background(), bson.M{
		"_id":   objID,
		"name":  "Test User",
		"email": "test@example.com",
	})
	assert.NoError(t, err)
	userID := objID.Hex()
	repo := repository.NewMongoUserRepository(db)
	uc := usecase.NewUserUseCase(repo, nil)

	tests := []struct {
		name        string
		inputID     string
		wantName    string
		expectError bool
	}{
		{
			name:        "valid user found",
			inputID:     userID,
			wantName:    "Test User",
			expectError: false,
		},
		{
			name:        "invalid hex ID",
			inputID:     "invalid-hex",
			wantName:    "",
			expectError: true,
		},
		{
			name:        "user not found",
			inputID:     primitive.NewObjectID().Hex(),
			wantName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := uc.GetByID(tt.inputID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, foundUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, foundUser)
				assert.Equal(t, tt.wantName, foundUser.Name)
			}
		})
	}
}
