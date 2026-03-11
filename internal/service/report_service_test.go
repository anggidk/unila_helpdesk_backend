package service

import (
	"math"
	"testing"
	"time"

	"gorm.io/datatypes"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
)

func TestPeriodStart(t *testing.T) {
	tests := []struct {
		name   string
		value  time.Time
		unit   string
		expect time.Time
	}{
		{
			name:   "daily start uses wib-localized midnight",
			value:  time.Date(2026, 3, 10, 23, 30, 0, 0, time.UTC),
			unit:   "daily",
			expect: time.Date(2026, 3, 11, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "weekly start aligns to monday",
			value:  time.Date(2026, 3, 11, 14, 15, 0, 0, reportLocationWIB),
			unit:   "weekly",
			expect: time.Date(2026, 3, 9, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "yearly start uses first day of year",
			value:  time.Date(2026, 8, 19, 8, 0, 0, 0, reportLocationWIB),
			unit:   "yearly",
			expect: time.Date(2026, 1, 1, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "default start uses first day of month",
			value:  time.Date(2026, 8, 19, 8, 0, 0, 0, reportLocationWIB),
			unit:   "quarterly",
			expect: time.Date(2026, 8, 1, 0, 0, 0, 0, reportLocationWIB),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := periodStart(tc.value, tc.unit); !got.Equal(tc.expect) {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestAddPeriods(t *testing.T) {
	base := time.Date(2026, 3, 9, 0, 0, 0, 0, reportLocationWIB)

	tests := []struct {
		name   string
		value  time.Time
		unit   string
		count  int
		expect time.Time
	}{
		{
			name:   "daily adds calendar days",
			value:  base,
			unit:   "daily",
			count:  2,
			expect: time.Date(2026, 3, 11, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "weekly subtracts seven day blocks",
			value:  base,
			unit:   "weekly",
			count:  -1,
			expect: time.Date(2026, 3, 2, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "yearly adds years",
			value:  base,
			unit:   "yearly",
			count:  1,
			expect: time.Date(2027, 3, 9, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:   "default adds months",
			value:  time.Date(2026, 1, 1, 0, 0, 0, 0, reportLocationWIB),
			unit:   "monthly",
			count:  3,
			expect: time.Date(2026, 4, 1, 0, 0, 0, 0, reportLocationWIB),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := addPeriods(tc.value, tc.unit, tc.count); !got.Equal(tc.expect) {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestPeriodRange(t *testing.T) {
	now := func() time.Time {
		return time.Date(2026, 3, 9, 10, 15, 0, 0, reportLocationWIB)
	}

	tests := []struct {
		name        string
		period      string
		periods     int
		expectStart time.Time
		expectEnd   time.Time
	}{
		{
			name:        "daily range uses day buckets",
			period:      "daily",
			periods:     3,
			expectStart: time.Date(2026, 3, 7, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "monthly range spans requested windows",
			period:      "monthly",
			periods:     2,
			expectStart: time.Date(2026, 2, 1, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 4, 1, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "invalid period and non positive count use monthly defaults",
			period:      "quarterly",
			periods:     0,
			expectStart: time.Date(2025, 11, 1, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 4, 1, 0, 0, 0, 0, reportLocationWIB),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotEnd := periodRange(tc.period, tc.periods, now)
			if !gotStart.Equal(tc.expectStart) {
				t.Fatalf("expected start %v, got %v", tc.expectStart, gotStart)
			}
			if !gotEnd.Equal(tc.expectEnd) {
				t.Fatalf("expected end %v, got %v", tc.expectEnd, gotEnd)
			}
		})
	}
}

func TestRollingReportRange(t *testing.T) {
	now := func() time.Time {
		return time.Date(2026, 3, 9, 10, 15, 0, 0, reportLocationWIB)
	}

	tests := []struct {
		name        string
		period      string
		expectStart time.Time
		expectEnd   time.Time
	}{
		{
			name:        "daily window",
			period:      "daily",
			expectStart: time.Date(2026, 3, 9, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "weekly window",
			period:      "weekly",
			expectStart: time.Date(2026, 3, 3, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "monthly window is default",
			period:      "monthly",
			expectStart: time.Date(2026, 2, 8, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "yearly window",
			period:      "yearly",
			expectStart: time.Date(2025, 3, 10, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
		{
			name:        "unknown period falls back to monthly",
			period:      "invalid",
			expectStart: time.Date(2026, 2, 8, 0, 0, 0, 0, reportLocationWIB),
			expectEnd:   time.Date(2026, 3, 10, 0, 0, 0, 0, reportLocationWIB),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotStart, gotEnd := rollingReportRange(tc.period, now)
			if !gotStart.Equal(tc.expectStart) {
				t.Fatalf("expected start %v, got %v", tc.expectStart, gotStart)
			}
			if !gotEnd.Equal(tc.expectEnd) {
				t.Fatalf("expected end %v, got %v", tc.expectEnd, gotEnd)
			}
		})
	}
}

func TestScoreFromResponseItem(t *testing.T) {
	scoreValue := 4.25

	tests := []struct {
		name         string
		item         domain.SurveyResponseItem
		questionType domain.SurveyQuestionType
		wantScore    float64
		wantOK       bool
	}{
		{
			name: "score value overrides answer payload",
			item: domain.SurveyResponseItem{
				ScoreValue:  &scoreValue,
				AnswerValue: datatypes.JSON(`{invalid`),
			},
			questionType: domain.QuestionLikert5,
			wantScore:    4.25,
			wantOK:       true,
		},
		{
			name: "empty answer without score value is invalid",
			item: domain.SurveyResponseItem{
				AnswerValue: datatypes.JSON{},
			},
			questionType: domain.QuestionYesNo,
			wantScore:    0,
			wantOK:       false,
		},
		{
			name: "invalid json returns false",
			item: domain.SurveyResponseItem{
				AnswerValue: datatypes.JSON(`{invalid`),
			},
			questionType: domain.QuestionYesNo,
			wantScore:    0,
			wantOK:       false,
		},
		{
			name: "boolean json is scored for yes no question",
			item: domain.SurveyResponseItem{
				AnswerValue: datatypes.JSON(`true`),
			},
			questionType: domain.QuestionYesNo,
			wantScore:    5,
			wantOK:       true,
		},
		{
			name: "numeric json string is scored for likert question",
			item: domain.SurveyResponseItem{
				AnswerValue: datatypes.JSON(`"2"`),
			},
			questionType: domain.QuestionLikert4,
			wantScore:    2.333333333333333,
			wantOK:       true,
		},
		{
			name: "unsupported question type returns false",
			item: domain.SurveyResponseItem{
				AnswerValue: datatypes.JSON(`"free form"`),
			},
			questionType: domain.QuestionText,
			wantScore:    0,
			wantOK:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := scoreFromResponseItem(tc.item, tc.questionType)
			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if !almostEqualReportScore(got, tc.wantScore) {
				t.Fatalf("expected score=%v, got %v", tc.wantScore, got)
			}
		})
	}
}

func TestBuildEntityPreferenceOverview(t *testing.T) {
	rows := []repository.EntityCategoryTotalRow{
		{Entity: domain.EntityMahasiswa, CategoryID: "1", Total: 6},
		{Entity: domain.EntityMahasiswa, CategoryID: "2", Total: 3},
		{Entity: domain.EntityDosen, CategoryID: "2", Total: 4},
		{Entity: domain.EntityDosen, CategoryID: "3", Total: 4},
		{Entity: "", CategoryID: "9", Total: 3},
	}
	categories := map[string]string{
		"1": "Akademik",
		"2": "Keuangan",
		"3": "Jaringan",
		"9": "Umum",
	}

	got := buildEntityPreferenceOverview(rows, categories)
	if len(got) != 4 {
		t.Fatalf("expected 4 ordered entity preferences, got %d", len(got))
	}

	expected := []domain.SatisfactionEntityPreferenceDTO{
		{
			Entity:    domain.EntityMahasiswa,
			Category:  "Akademik",
			Responses: 6,
			Share:     66.7,
		},
		{
			Entity:    domain.EntityDosen,
			Category:  "Jaringan",
			Responses: 4,
			Share:     50,
		},
		{
			Entity: domain.EntityTendik,
		},
		{
			Entity:    domain.EntityLainnya,
			Category:  "Umum",
			Responses: 3,
			Share:     100,
		},
	}

	for index := range expected {
		if got[index].Entity != expected[index].Entity {
			t.Fatalf("expected entity %q at index %d, got %q", expected[index].Entity, index, got[index].Entity)
		}
		if got[index].Category != expected[index].Category {
			t.Fatalf("expected category %q at index %d, got %q", expected[index].Category, index, got[index].Category)
		}
		if got[index].Responses != expected[index].Responses {
			t.Fatalf("expected responses %d at index %d, got %d", expected[index].Responses, index, got[index].Responses)
		}
		if !almostEqualReportScore(got[index].Share, expected[index].Share) {
			t.Fatalf("expected share %v at index %d, got %v", expected[index].Share, index, got[index].Share)
		}
	}
}

func almostEqualReportScore(left float64, right float64) bool {
	return math.Abs(left-right) < 1e-9
}
