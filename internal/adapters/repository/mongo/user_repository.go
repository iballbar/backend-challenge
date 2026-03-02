package mongo

import (
	"context"
	"errors"
	"time"

	"backend-challenge/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName string) *UserRepository {
	return &UserRepository{
		collection: client.Database(dbName).Collection("users"),
	}
}

type userRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"createdAt"`
}

func (r *UserRepository) Create(ctx context.Context, input domain.CreateUser) (domain.User, error) {
	record := userRecord{
		ID:        primitive.NewObjectID(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password,
		CreatedAt: time.Now().UTC(),
	}
	_, err := r.collection.InsertOne(ctx, record)
	if err != nil {
		return domain.User{}, err
	}
	return toUser(record), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	var record userRecord
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return toUser(record), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var record userRecord
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return toUser(record), nil
}

func (r *UserRepository) List(ctx context.Context, page, limit int) ([]domain.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	pageNumber := int64(page)
	pageSize := int64(limit)
	skipAmount := (pageNumber - 1) * pageSize
	findOptions := options.Find()
	findOptions.SetSkip(skipAmount)
	findOptions.SetLimit(pageSize)

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var results []domain.User
	for cursor.Next(ctx) {
		var record userRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, 0, err
		}
		results = append(results, toUser(record))
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, input domain.UpdateUser) (domain.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	update := bson.M{}
	if input.Name != nil {
		update["name"] = *input.Name
	}
	if input.Email != nil {
		update["email"] = *input.Email
	}
	if len(update) == 0 {
		return r.GetByID(ctx, id)
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var record userRecord
	err = r.collection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, bson.M{"$set": update}, opts).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	return toUser(record), nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrNotFound
	}
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.D{})
}

func toUser(record userRecord) domain.User {
	return domain.User{
		ID:        record.ID.Hex(),
		Name:      record.Name,
		Email:     record.Email,
		Password:  record.Password,
		CreatedAt: record.CreatedAt,
	}
}
