package services

import (
	"context"
	"testing"

	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports/mocks"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	type args struct {
		input domain.CreateUser
	}
	tests := []struct {
		name    string
		args    args
		mock    func(repo *mocks.MockUserRepository)
		wantErr error
	}{
		{
			name: "Success",
			args: args{
				input: domain.CreateUser{
					Name:     "Alice",
					Email:    "alice@example.com",
					Password: "secret",
				},
			},
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "alice@example.com").Return(domain.User{}, domain.ErrNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, input domain.CreateUser) (domain.User, error) {
					if input.Email != "alice@example.com" {
						t.Errorf("unexpected email: %s", input.Email)
					}
					if bcrypt.CompareHashAndPassword([]byte(input.Password), []byte("secret")) != nil {
						t.Errorf("password hash mismatch")
					}
					return domain.User{
						ID:    "1",
						Name:  input.Name,
						Email: input.Email,
					}, nil
				})
			},
			wantErr: nil,
		},
		{
			name: "Invalid Email",
			args: args{
				input: domain.CreateUser{
					Name:     "Alice",
					Email:    "invalid",
					Password: "secret",
				},
			},
			mock:    func(repo *mocks.MockUserRepository) {},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name: "Email Already Exists",
			args: args{
				input: domain.CreateUser{
					Name:     "Alice",
					Email:    "exists@example.com",
					Password: "secret",
				},
			},
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "exists@example.com").Return(domain.User{ID: "1"}, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			tt.mock(repo)
			s := NewService(repo)
			_, err := s.Create(context.Background(), tt.args.input)
			if err != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Authenticate(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		email    string
		password string
		mock     func(repo *mocks.MockUserRepository)
		wantErr  error
	}{
		{
			name:     "Success",
			email:    "alice@example.com",
			password: "secret",
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "alice@example.com").Return(domain.User{
					ID:       "1",
					Email:    "alice@example.com",
					Password: string(hashedPassword),
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "User Not Found",
			email:    "notfound@example.com",
			password: "secret",
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "notfound@example.com").Return(domain.User{}, domain.ErrNotFound)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:     "Wrong Password",
			email:    "alice@example.com",
			password: "wrong",
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "alice@example.com").Return(domain.User{
					ID:       "1",
					Email:    "alice@example.com",
					Password: string(hashedPassword),
				}, nil)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			tt.mock(repo)
			s := NewService(repo)
			_, err := s.Authenticate(context.Background(), tt.email, tt.password)
			if err != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mock    func(repo *mocks.MockUserRepository)
		wantErr error
	}{
		{
			name: "Success",
			id:   "1",
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), "1").Return(domain.User{ID: "1", Name: "Alice"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Not Found",
			id:   "2",
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), "2").Return(domain.User{}, nil)
			},
			wantErr: domain.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			tt.mock(repo)
			s := NewService(repo)
			_, err := s.Get(context.Background(), tt.id)
			if err != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	newName := "Alice Updated"
	newEmail := "alice.new@example.com"
	invalidEmail := "invalid"

	tests := []struct {
		name    string
		id      string
		input   domain.UpdateUser
		mock    func(repo *mocks.MockUserRepository)
		wantErr error
	}{
		{
			name: "Success",
			id:   "1",
			input: domain.UpdateUser{
				Name:  &newName,
				Email: &newEmail,
			},
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "alice.new@example.com").Return(domain.User{}, nil)
				repo.EXPECT().Update(gomock.Any(), "1", gomock.Any()).Return(domain.User{ID: "1", Name: newName, Email: newEmail}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Invalid Email",
			id:   "1",
			input: domain.UpdateUser{
				Email: &invalidEmail,
			},
			mock:    func(repo *mocks.MockUserRepository) {},
			wantErr: domain.ErrInvalidEmail,
		},
		{
			name: "Email Conflict",
			id:   "1",
			input: domain.UpdateUser{
				Email: &newEmail,
			},
			mock: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByEmail(gomock.Any(), "alice.new@example.com").Return(domain.User{ID: "2", Email: "alice.new@example.com"}, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockUserRepository(ctrl)
			tt.mock(repo)
			s := NewService(repo)
			_, err := s.Update(context.Background(), tt.id, tt.input)
			if err != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	s := NewService(repo)
	ctx := context.Background()

	users := []domain.User{{ID: "1", Name: "Alice"}}
	repo.EXPECT().List(ctx, 1, 10).Return(users, int64(1), nil)

	got, total, err := s.List(ctx, 1, 10)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("List() got = %v, want 1 user", len(got))
	}
	if total != 1 {
		t.Errorf("List() total = %v, want 1", total)
	}
}

func TestService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	s := NewService(repo)
	ctx := context.Background()

	repo.EXPECT().Delete(ctx, "1").Return(nil)

	if err := s.Delete(ctx, "1"); err != nil {
		t.Errorf("Delete() error = %v", err)
	}
}

func TestService_Count(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepository(ctrl)
	s := NewService(repo)
	ctx := context.Background()

	repo.EXPECT().Count(ctx).Return(int64(5), nil)

	got, err := s.Count(ctx)
	if err != nil {
		t.Errorf("Count() error = %v", err)
	}
	if got != 5 {
		t.Errorf("Count() got = %v, want 5", got)
	}
}
