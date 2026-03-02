package integration_test

import (
	"backend-challenge/internal/core/domain"
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// lotteryTicketDoc mirrors the internal lotteryRecord BSON layout.
type lotteryTicketDoc struct {
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

// seedTickets inserts lottery tickets directly via the raw mongo client.
func seedTickets(t *testing.T, docs []lotteryTicketDoc) {
	t.Helper()
	ctx := context.Background()
	coll := mongoClient.Database(testDBName).Collection("lottery_tickets")
	raw := make([]any, len(docs))
	for i, d := range docs {
		if d.ID.IsZero() {
			d.ID = primitive.NewObjectID()
		}
		raw[i] = d
	}
	if _, err := coll.InsertMany(ctx, raw); err != nil {
		t.Fatalf("seedTickets: %v", err)
	}
}

// reserveTickets marks tickets as reserved until a future time.
func reserveTickets(t *testing.T, ids []primitive.ObjectID, until time.Time) {
	t.Helper()
	ctx := context.Background()
	coll := mongoClient.Database(testDBName).Collection("lottery_tickets")
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "reservedUntil", Value: until}}}}
	if _, err := coll.UpdateMany(ctx, filter, update); err != nil {
		t.Fatalf("reserveTickets: %v", err)
	}
}

// makeTicket builds a lotteryTicketDoc from a 6-digit number string.
func makeTicket(number string, set int) lotteryTicketDoc {
	if len(number) != 6 {
		panic("number must be 6 chars")
	}
	return lotteryTicketDoc{
		ID:     primitive.NewObjectID(),
		Number: number,
		Set:    set,
		Rand:   rand.Float64(),
		D1:     int8(number[0] - '0'),
		D2:     int8(number[1] - '0'),
		D3:     int8(number[2] - '0'),
		D4:     int8(number[3] - '0'),
		D5:     int8(number[4] - '0'),
		D6:     int8(number[5] - '0'),
	}
}

func TestLotteryRepository_Search(t *testing.T) {
	if os.Getenv("INT") != "1" {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		pattern  string
		limit    int
		validate func(t *testing.T, results []domain.LotteryTicket, count int64, err error)
	}{
		{
			name: "Wildcard matches all available",
			setup: func(t *testing.T) {
				seedTickets(t, []lotteryTicketDoc{
					makeTicket("123456", 1),
					makeTicket("111111", 1),
				})
			},
			pattern: "******",
			limit:   5,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 2)
			},
		},
		{
			name: "Specific digits matching",
			setup: func(t *testing.T) {
				seedTickets(t, []lotteryTicketDoc{
					makeTicket("100005", 1),
					makeTicket("200006", 1),
				})
			},
			pattern: "1****5",
			limit:   5,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "100005", results[0].Number)
			},
		},
		{
			name: "No results for non-matching pattern",
			setup: func(t *testing.T) {
				seedTickets(t, []lotteryTicketDoc{makeTicket("123456", 1)})
			},
			pattern: "99999*",
			limit:   5,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				assert.Empty(t, results)
			},
		},
		{
			name: "Reserved tickets are excluded",
			setup: func(t *testing.T) {
				res := makeTicket("111111", 1)
				avail := makeTicket("222222", 1)
				seedTickets(t, []lotteryTicketDoc{res, avail})
				reserveTickets(t, []primitive.ObjectID{res.ID}, time.Now().Add(10*time.Minute))
			},
			pattern: "******",
			limit:   10,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "222222", results[0].Number)
			},
		},
		{
			name: "Expired reservations are returned",
			setup: func(t *testing.T) {
				exp := makeTicket("333333", 1)
				seedTickets(t, []lotteryTicketDoc{exp})
				reserveTickets(t, []primitive.ObjectID{exp.ID}, time.Now().Add(-1*time.Minute))
			},
			pattern: "3333**",
			limit:   5,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "333333", results[0].Number)
			},
		},
		{
			name: "Limit is respected",
			setup: func(t *testing.T) {
				var docs []lotteryTicketDoc
				for i := 0; i < 10; i++ {
					docs = append(docs, makeTicket(fmt.Sprintf("1%05d", i), 1))
				}
				seedTickets(t, docs)
			},
			pattern: "1*****",
			limit:   4,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				require.NoError(t, err)
				assert.LessOrEqual(t, len(results), 4)
			},
		},
		{
			name:    "Invalid pattern returns error",
			setup:   func(t *testing.T) {},
			pattern: "123", // too short
			limit:   1,
			validate: func(t *testing.T, results []domain.LotteryTicket, count int64, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			truncateCollection(t, "lottery_tickets")
			tc.setup(t)
			results, count, err := lotteryRepo.Search(ctx, tc.pattern, tc.limit)
			tc.validate(t, results, count, err)
		})
	}
}
