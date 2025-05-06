package repository

import (
	"context"
	"github.com/facelessEmptiness/user_service/internal/domain"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUserRepo struct {
	coll *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) UserRepository {
	return &mongoUserRepo{coll: db.Collection("users")}
}

func (r *mongoUserRepo) Create(user *domain.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doc := bson.M{
		"name":       user.Name,
		"email":      user.Email,
		"password":   user.Password,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	result, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	oid := result.InsertedID.(primitive.ObjectID).Hex()
	return oid, nil
}

func (r *mongoUserRepo) GetByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var u domain.User
	if err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *mongoUserRepo) GetByID(id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var u domain.User
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *mongoUserRepo) Update(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"password":   user.Password,
			"updated_at": time.Now(),
		},
	}

	_, err = r.coll.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)
	return err
}

func (r *mongoUserRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *mongoUserRepo) List(page, limit int64) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Calculate skip value for pagination
	skip := (page - 1) * limit

	// Setup options for pagination and sorting
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at descending

	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}
