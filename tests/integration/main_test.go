package integration_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	httpAdapter "backend-challenge/internal/adapters/http"
	mongorepo "backend-challenge/internal/adapters/repository/mongo"
	tokenAdapter "backend-challenge/internal/adapters/token"
	"backend-challenge/internal/core/services"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDBName    = "backend_challenge_test"
	testJWTSecret = "test-secret-key-123"
)

var (
	mongoClient *mongo.Client
	userRepo    *mongorepo.UserRepository
	lotteryRepo *mongorepo.LotteryRepository
	testRouter  *gin.Engine
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	if os.Getenv("INT") != "1" {
		slog.Info("INT is not set, skipping integration tests")
		return m.Run()
	}

	ctx := context.Background()

	// Disable Ryuk (the testcontainers reaper) — required on Windows Docker Desktop.
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongo:8")
	if err != nil {
		slog.Error("failed to start MongoDB container", "error", err)
		return 1
	}
	defer func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			slog.Error("failed to terminate MongoDB container", "error", err)
		}
	}()

	// Get connection URI from container
	uri, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		slog.Error("failed to get MongoDB URI", "error", err)
		return 1
	}

	// Connect mongo client
	clientOpts := options.Client().ApplyURI(uri)
	mongoClient, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		slog.Error("failed to connect to MongoDB", "error", err)
		return 1
	}
	defer mongoClient.Disconnect(ctx)

	// Ping to confirm connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		slog.Error("failed to ping MongoDB", "error", err)
		return 1
	}

	// Bootstrap indexes
	bootstrapIndexes(ctx)

	// Initialise repositories
	userRepo = mongorepo.NewUserRepository(mongoClient, testDBName)
	lotteryRepo = mongorepo.NewLotteryRepository(mongoClient, testDBName, 5*time.Minute)

	// Initialise services
	userService := services.NewService(userRepo)
	lotteryService := services.NewLotteryService(lotteryRepo)
	tokenProvider := tokenAdapter.NewTokenService(testJWTSecret)

	// Initialise handlers
	userHandler := httpAdapter.NewUserHandler(userService, tokenProvider)
	lotteryHandler := httpAdapter.NewLotteryHandler(lotteryService)

	// Initialise router
	gin.SetMode(gin.TestMode)
	testRouter = httpAdapter.NewRouter(userHandler, lotteryHandler, tokenProvider)

	return m.Run()
}

// bootstrapIndexes creates the same indexes as scripts/mongo-init/init.js.
func bootstrapIndexes(ctx context.Context) {
	db := mongoClient.Database(testDBName)

	// Users – unique email
	db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_email"),
	})

	lc := db.Collection("lottery_tickets")
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "number", Value: 1}, {Key: "set", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_number_set"),
		},
		{Keys: bson.D{{Key: "reservedUntil", Value: 1}}, Options: options.Index().SetName("reserved_until")},
		{Keys: bson.D{{Key: "d1", Value: 1}}, Options: options.Index().SetName("digit1")},
		{Keys: bson.D{{Key: "d2", Value: 1}}, Options: options.Index().SetName("digit2")},
		{Keys: bson.D{{Key: "d3", Value: 1}}, Options: options.Index().SetName("digit3")},
		{Keys: bson.D{{Key: "d4", Value: 1}}, Options: options.Index().SetName("digit4")},
		{Keys: bson.D{{Key: "d5", Value: 1}}, Options: options.Index().SetName("digit5")},
		{Keys: bson.D{{Key: "d6", Value: 1}}, Options: options.Index().SetName("digit6")},
		{Keys: bson.D{{Key: "rand", Value: 1}}, Options: options.Index().SetName("rand")},
		{
			Keys:    bson.D{{Key: "d1", Value: 1}, {Key: "d2", Value: 1}, {Key: "d3", Value: 1}, {Key: "d4", Value: 1}, {Key: "d5", Value: 1}, {Key: "d6", Value: 1}, {Key: "rand", Value: 1}},
			Options: options.Index().SetName("digits_all"),
		},
	}
	lc.Indexes().CreateMany(ctx, models)
}

// truncateCollection drops all documents in the given collection for test isolation.
func truncateCollection(t *testing.T, name string) {
	t.Helper()
	ctx := context.Background()
	if _, err := mongoClient.Database(testDBName).Collection(name).DeleteMany(ctx, bson.D{}); err != nil {
		t.Fatalf("truncateCollection %q: %v", name, err)
	}
}
