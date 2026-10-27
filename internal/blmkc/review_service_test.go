package blmkc

import (
	"errors"
	"testing"
)

func TestReviewCaseRejectsInvalidInputBeforeRepository(t *testing.T) {
	service := NewService(nil)
	tests := []struct {
		name, decision, notes string
		id, userID            uint64
	}{
		{"missing case", "verified", "", 0, 1},
		{"missing reviewer", "verified", "", 1, 0},
		{"unknown decision", "approve", "", 1, 1},
		{"revision without notes", "revision", "", 1, 1},
		{"rejection without notes", "rejected", "   ", 1, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := service.ReviewCase(t.Context(), test.id, test.userID, test.decision, test.notes); !errors.Is(err, ErrInvalid) {
				t.Fatalf("got %v, want ErrInvalid", err)
			}
		})
	}
}
