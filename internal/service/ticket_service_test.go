package service

import (
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

func TestTicketOwnedByUser(t *testing.T) {
	strPtr := func(value string) *string { return &value }

	tests := []struct {
		name   string
		ticket domain.Ticket
		user   domain.User
		expect bool
	}{
		{
			name: "exact match",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "registered.user", ID: "12345"},
			expect: true,
		},
		{
			name: "username is case insensitive",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "REGISTERED.USER", ID: "12345"},
			expect: true,
		},
		{
			name: "user username and id are trimmed",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "  registered.user ", ID: " 12345 "},
			expect: true,
		},
		{
			name: "mismatched username is rejected",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "other.user", ID: "12345"},
			expect: false,
		},
		{
			name: "mismatched number id is rejected",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "registered.user", ID: "67890"},
			expect: false,
		},
		{
			name: "nil username is rejected",
			ticket: domain.Ticket{
				Username: nil,
				NumberID: strPtr("12345"),
			},
			user:   domain.User{Username: "registered.user", ID: "12345"},
			expect: false,
		},
		{
			name: "nil number id is rejected",
			ticket: domain.Ticket{
				Username: strPtr("registered.user"),
				NumberID: nil,
			},
			user:   domain.User{Username: "registered.user", ID: "12345"},
			expect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ticketOwnedByUser(tc.ticket, tc.user); got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}
