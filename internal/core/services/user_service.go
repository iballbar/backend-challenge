package services

import (
	"context"
	"regexp"
	"strings"
	"time"

	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo ports.UserRepository
}

func NewService(repo ports.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) Create(ctx context.Context, input domain.CreateUser) (domain.User, error) {
	if !isValidEmail(input.Email) {
		return domain.User{}, domain.ErrInvalidEmail
	}
	existing, err := s.repo.GetByEmail(ctx, strings.ToLower(input.Email))
	if err == nil && existing.ID != "" {
		return domain.User{}, domain.ErrEmailAlreadyExists
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	input.Email = strings.ToLower(input.Email)
	input.Password = string(hashed)
	return s.repo.Create(ctx, input)
}

func (s *Service) Get(ctx context.Context, id string) (domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	if user.ID == "" {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (s *Service) List(ctx context.Context, page, limit int) ([]domain.User, int64, error) {
	return s.repo.List(ctx, page, limit)
}

func (s *Service) Update(ctx context.Context, id string, input domain.UpdateUser) (domain.User, error) {
	if input.Email != nil {
		if !isValidEmail(*input.Email) {
			return domain.User{}, domain.ErrInvalidEmail
		}
		email := strings.ToLower(*input.Email)
		input.Email = &email
		existing, err := s.repo.GetByEmail(ctx, email)
		if err == nil && existing.ID != "" && existing.ID != id {
			return domain.User{}, domain.ErrEmailAlreadyExists
		}
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

func NewUser(name, email, hashedPassword string) domain.User {
	return domain.User{
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
	}
}
