package mongo

import (
	"testing"

	"backend-challenge/internal/core/domain"
)

func TestPatternFilter(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		wantLen  int
		wantKeys []string
		wantVals []int8
		wantErr  error
	}{
		{
			name:     "Mixed pattern",
			pattern:  "1****5",
			wantLen:  2,
			wantKeys: []string{"d1", "d6"},
			wantVals: []int8{1, 5},
			wantErr:  nil,
		},
		{
			name:    "All wildcards",
			pattern: "******",
			wantLen: 0,
			wantErr: nil,
		},
		{
			name:     "All digits",
			pattern:  "123456",
			wantLen:  6,
			wantKeys: []string{"d1", "d2", "d3", "d4", "d5", "d6"},
			wantVals: []int8{1, 2, 3, 4, 5, 6},
			wantErr:  nil,
		},
		{
			name:    "Invalid length - short",
			pattern: "123",
			wantErr: domain.ErrInvalidPattern,
		},
		{
			name:    "Invalid length - long",
			pattern: "1234567",
			wantErr: domain.ErrInvalidPattern,
		},
		{
			name:    "Invalid characters",
			pattern: "123a45",
			wantErr: domain.ErrInvalidPattern,
		},
		{
			name:    "Empty string",
			pattern: "",
			wantErr: domain.ErrInvalidPattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := patternFilter(tt.pattern)
			if err != tt.wantErr {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if err != nil {
				return
			}
			if len(filter) != tt.wantLen {
				t.Fatalf("expected %d filter elements, got %d", tt.wantLen, len(filter))
			}
			for i, key := range tt.wantKeys {
				if filter[i].Key != key || filter[i].Value.(int8) != tt.wantVals[i] {
					t.Fatalf("expected %s=%v, got %s=%v", key, tt.wantVals[i], filter[i].Key, filter[i].Value)
				}
			}
		})
	}
}
