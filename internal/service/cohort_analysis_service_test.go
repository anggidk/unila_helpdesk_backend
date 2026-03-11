package service

import (
	"testing"
	"time"
)

func TestPeriodDiff(t *testing.T) {
	tests := []struct {
		name   string
		start  time.Time
		end    time.Time
		unit   string
		expect int
	}{
		{
			name:   "daily difference",
			start:  time.Date(2026, 3, 1, 10, 0, 0, 0, reportLocationWIB),
			end:    time.Date(2026, 3, 3, 9, 0, 0, 0, reportLocationWIB),
			unit:   "daily",
			expect: 2,
		},
		{
			name:   "weekly difference uses monday alignment",
			start:  time.Date(2026, 3, 10, 10, 0, 0, 0, reportLocationWIB),
			end:    time.Date(2026, 3, 16, 8, 0, 0, 0, reportLocationWIB),
			unit:   "weekly",
			expect: 1,
		},
		{
			name:   "monthly difference across year boundary",
			start:  time.Date(2025, 11, 15, 10, 0, 0, 0, reportLocationWIB),
			end:    time.Date(2026, 2, 2, 9, 0, 0, 0, reportLocationWIB),
			unit:   "monthly",
			expect: 3,
		},
		{
			name:   "yearly difference",
			start:  time.Date(2024, 6, 1, 10, 0, 0, 0, reportLocationWIB),
			end:    time.Date(2026, 1, 10, 9, 0, 0, 0, reportLocationWIB),
			unit:   "yearly",
			expect: 2,
		},
		{
			name:   "negative difference when end is before start",
			start:  time.Date(2026, 3, 5, 10, 0, 0, 0, reportLocationWIB),
			end:    time.Date(2026, 3, 4, 10, 0, 0, 0, reportLocationWIB),
			unit:   "daily",
			expect: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := periodDiff(tc.start, tc.end, tc.unit); got != tc.expect {
				t.Fatalf("expected %d, got %d", tc.expect, got)
			}
		})
	}
}
