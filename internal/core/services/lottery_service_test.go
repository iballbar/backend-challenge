package services

import (
	"context"
	"errors"
	"testing"

	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports/mocks"

	"go.uber.org/mock/gomock"
)

func TestLotteryService_Search(t *testing.T) {
	type fields struct {
		repo *mocks.MockLotteryRepository
	}
	type args struct {
		pattern string
		limit   int
	}
	tests := []struct {
		name    string
		args    args
		mock    func(f fields)
		want    []domain.LotteryTicket
		wantRem int64
		wantErr error
	}{
		{
			name: "Success - tickets found",
			args: args{
				pattern: "123***",
				limit:   10,
			},
			mock: func(f fields) {
				f.repo.EXPECT().
					Search(gomock.Any(), "123***", 10).
					Return([]domain.LotteryTicket{{ID: "1", Number: "123456", Set: 1}}, int64(5), nil)
			},
			want:    []domain.LotteryTicket{{ID: "1", Number: "123456", Set: 1}},
			wantRem: 5,
			wantErr: nil,
		},
		{
			name: "Success - no tickets found",
			args: args{
				pattern: "999999",
				limit:   10,
			},
			mock: func(f fields) {
				f.repo.EXPECT().
					Search(gomock.Any(), "999999", 10).
					Return([]domain.LotteryTicket{}, int64(0), nil)
			},
			want:    nil,
			wantRem: 0,
			wantErr: domain.ErrNoTicketsAvailable,
		},
		{
			name: "Error - repository failure",
			args: args{
				pattern: "******",
				limit:   5,
			},
			mock: func(f fields) {
				f.repo.EXPECT().
					Search(gomock.Any(), "******", 5).
					Return(nil, int64(0), errors.New("db down"))
			},
			want:    nil,
			wantRem: 0,
			wantErr: errors.New("db down"),
		},
		{
			name: "Success - limit normalization",
			args: args{
				pattern: "111***",
				limit:   -5,
			},
			mock: func(f fields) {
				f.repo.EXPECT().
					Search(gomock.Any(), "111***", 1).
					Return([]domain.LotteryTicket{{ID: "2", Number: "111222", Set: 1}}, int64(10), nil)
			},
			want:    []domain.LotteryTicket{{ID: "2", Number: "111222", Set: 1}},
			wantRem: 10,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				repo: mocks.NewMockLotteryRepository(ctrl),
			}
			tt.mock(f)

			s := NewLotteryService(f.repo)
			got, rem, err := s.Search(context.Background(), tt.args.pattern, tt.args.limit)

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("Search() unexpected error = %v", err)
				return
			}

			if rem != tt.wantRem {
				t.Errorf("Search() remaining = %v, want %v", rem, tt.wantRem)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Search() got count = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}
