package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"backend-challenge/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	if os.Getenv("INT") != "1" {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		truncateCollection(t, "users")
		input := domain.CreateUser{Name: "Alice", Email: "alice@example.com", Password: "hashed"}
		user, err := userRepo.Create(ctx, input)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "Alice", user.Name)
	})

	t.Run("GetByID", func(t *testing.T) {
		truncateCollection(t, "users")
		created, _ := userRepo.Create(ctx, domain.CreateUser{Name: "Bob", Email: "bob@example.com", Password: "hashed"})

		t.Run("Success", func(t *testing.T) {
			got, err := userRepo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, created.ID, got.ID)
		})

		t.Run("NotFound", func(t *testing.T) {
			_, err := userRepo.GetByID(ctx, "000000000000000000000000")
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	})

	t.Run("GetByEmail", func(t *testing.T) {
		truncateCollection(t, "users")
		userRepo.Create(ctx, domain.CreateUser{Name: "Carol", Email: "carol@example.com", Password: "hashed"})

		t.Run("Success", func(t *testing.T) {
			got, err := userRepo.GetByEmail(ctx, "carol@example.com")
			require.NoError(t, err)
			assert.Equal(t, "Carol", got.Name)
		})

		t.Run("NotFound", func(t *testing.T) {
			_, err := userRepo.GetByEmail(ctx, "nobody@example.com")
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	})

	t.Run("List", func(t *testing.T) {
		truncateCollection(t, "users")
		for i := 0; i < 5; i++ {
			email := "user" + strings.Repeat("a", i) + "@example.com"
			userRepo.Create(ctx, domain.CreateUser{Name: "User", Email: email, Password: "hashed"})
		}

		users, total, err := userRepo.List(ctx, 1, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, users, 3)

		users2, _, _ := userRepo.List(ctx, 2, 3)
		assert.Len(t, users2, 2)
	})

	t.Run("Update", func(t *testing.T) {
		truncateCollection(t, "users")
		created, _ := userRepo.Create(ctx, domain.CreateUser{Name: "Dave", Email: "dave@example.com", Password: "hashed"})

		t.Run("Success", func(t *testing.T) {
			newName := "David"
			updated, err := userRepo.Update(ctx, created.ID, domain.UpdateUser{Name: &newName})
			require.NoError(t, err)
			assert.Equal(t, "David", updated.Name)
		})

		t.Run("NotFound", func(t *testing.T) {
			newName := "Ghost"
			_, err := userRepo.Update(ctx, "000000000000000000000000", domain.UpdateUser{Name: &newName})
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		truncateCollection(t, "users")
		created, _ := userRepo.Create(ctx, domain.CreateUser{Name: "Eve", Email: "eve@example.com", Password: "hashed"})

		t.Run("Success", func(t *testing.T) {
			err := userRepo.Delete(ctx, created.ID)
			require.NoError(t, err)
			_, err = userRepo.GetByID(ctx, created.ID)
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})

		t.Run("NotFound", func(t *testing.T) {
			err := userRepo.Delete(ctx, "000000000000000000000000")
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	})

	t.Run("Count", func(t *testing.T) {
		truncateCollection(t, "users")
		count, _ := userRepo.Count(ctx)
		assert.Equal(t, int64(0), count)

		for i := 0; i < 3; i++ {
			email := "cnt" + strings.Repeat("x", i) + "@example.com"
			userRepo.Create(ctx, domain.CreateUser{Name: "U", Email: email, Password: "p"})
		}
		count, _ = userRepo.Count(ctx)
		assert.Equal(t, int64(3), count)
	})
}
