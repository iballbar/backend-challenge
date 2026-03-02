package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend-challenge/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LotteryRepository struct {
	collection     *mongo.Collection
	reservationTTL time.Duration
}

func NewLotteryRepository(client *mongo.Client, dbName string, ttl time.Duration) *LotteryRepository {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	repo := &LotteryRepository{
		collection:     client.Database(dbName).Collection("lottery_tickets"),
		reservationTTL: ttl,
	}

	return repo
}

type lotteryRecord struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Number        string             `bson:"number"`
	Set           int                `bson:"set"`
	Rand          float64            `bson:"rand"`
	D1            int8               `bson:"d1"`
	D2            int8               `bson:"d2"`
	D3            int8               `bson:"d3"`
	D4            int8               `bson:"d4"`
	D5            int8               `bson:"d5"`
	D6            int8               `bson:"d6"`
	ReservedUntil *time.Time         `bson:"reservedUntil,omitempty"`
}

func (r *LotteryRepository) Search(ctx context.Context, pattern string, limit int) ([]domain.LotteryTicket, int64, error) {
	if limit <= 0 {
		limit = 1
	}

	digitFilter, err := patternFilter(pattern)
	if err != nil {
		return []domain.LotteryTicket{}, 0, err
	}

	// Base availability filter - ticket must be unreserved or expired
	now := time.Now().UTC()
	availableFilter := bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{{Key: "reservedUntil", Value: nil}},
			bson.D{
				{Key: "reservedUntil", Value: bson.D{{Key: "$lte", Value: now}}},
			},
		},
	}

	filter := bson.D{availableFilter}
	filter = append(filter, digitFilter...)

	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(int64(limit + limit)). // set limit + buffer to avoid missing tickets
		SetSort(bson.D{{Key: "rand", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	var candidates []lotteryRecord
	if err = cursor.All(ctx, &candidates); err != nil {
		return nil, 0, err
	}

	if len(candidates) == 0 {
		return []domain.LotteryTicket{}, 0, nil
	}

	expired := now.Add(r.reservationTTL)
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "reservedUntil", Value: expired}}}}
	results := make([]domain.LotteryTicket, 0, limit)
	remaining := int64(len(candidates))
	for _, rec := range candidates {
		if len(results) >= limit {
			break
		}

		// Exact match allocate by _id (prevents concurrent race conditions on reservation)
		allocFilter := bson.D{
			{Key: "_id", Value: rec.ID},
			availableFilter,
		}

		ticket, ok, err := r.allocate(ctx, allocFilter, update)
		if err != nil {
			continue
		}
		if ok {
			results = append(results, ticket)
		}
		remaining--
	}

	return results, remaining, nil
}

func (r *LotteryRepository) allocate(ctx context.Context, filter bson.D, update bson.D) (domain.LotteryTicket, bool, error) {

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var rec lotteryRecord
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&rec)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.LotteryTicket{}, false, nil
	}
	if err != nil {
		return domain.LotteryTicket{}, false, err
	}

	return domain.LotteryTicket{
		ID:     rec.ID.Hex(),
		Number: rec.Number,
		Set:    rec.Set,
	}, true, nil
}

func patternFilter(pattern string) (bson.D, error) {
	if len(pattern) != 6 {
		return nil, domain.ErrInvalidPattern
	}

	filter := bson.D{}
	for pos := 0; pos < 6; pos++ {
		ch := pattern[pos]
		if ch == '*' {
			continue
		}
		if ch < '0' || ch > '9' {
			return nil, domain.ErrInvalidPattern
		}
		key := fmt.Sprintf("d%d", pos+1)
		filter = append(filter, bson.E{Key: key, Value: int8(ch - '0')})
	}
	return filter, nil
}
