package service

import (
	"testing"

	"unila_helpdesk_backend/internal/domain"
)

// ============================================================================
// Tests for DefaultCategories
// ============================================================================

func TestDefaultCategories_Count(t *testing.T) {
	categories := DefaultCategories()
	expectedCount := 8
	if len(categories) != expectedCount {
		t.Errorf("DefaultCategories() returned %d categories, want %d", len(categories), expectedCount)
	}
}

func TestDefaultCategories_IDs(t *testing.T) {
	categories := DefaultCategories()

	expectedIDs := []string{
		CategoryInternet,
		CategorySIAKAD,
		CategoryWebsite,
		CategorySistemInformasi,
		CategoryLainnya,
		CategoryGuestPassword,
		CategoryGuestSSORegistration,
		CategoryGuestEmailRegistration,
	}

	for i, expected := range expectedIDs {
		if categories[i].ID != expected {
			t.Errorf("Category[%d].ID = %q, want %q", i, categories[i].ID, expected)
		}
	}
}

func TestDefaultCategories_GuestAllowed(t *testing.T) {
	categories := DefaultCategories()

	// Map of ID -> expected GuestAllowed
	expected := map[string]bool{
		CategoryInternet:               false,
		CategorySIAKAD:                 false,
		CategoryWebsite:                false,
		CategorySistemInformasi:        false,
		CategoryLainnya:                false,
		CategoryGuestPassword:          true,
		CategoryGuestSSORegistration:   true,
		CategoryGuestEmailRegistration: true,
	}

	for _, cat := range categories {
		want, ok := expected[cat.ID]
		if !ok {
			t.Errorf("Unexpected category ID: %s", cat.ID)
			continue
		}
		if cat.GuestAllowed != want {
			t.Errorf("Category %s GuestAllowed = %v, want %v", cat.ID, cat.GuestAllowed, want)
		}
	}
}

func TestDefaultCategories_Names(t *testing.T) {
	categories := DefaultCategories()

	// Check that all categories have non-empty names
	for _, cat := range categories {
		if cat.Name == "" {
			t.Errorf("Category %s has empty name", cat.ID)
		}
	}
}

func TestDefaultCategories_SpecificNames(t *testing.T) {
	categories := DefaultCategories()

	names := make(map[string]string)
	for _, cat := range categories {
		names[cat.ID] = cat.Name
	}

	if names[CategoryInternet] != "Jaringan Internet" {
		t.Errorf("Internet category name = %q, want 'Jaringan Internet'", names[CategoryInternet])
	}
	if names[CategorySIAKAD] != "SIAKAD" {
		t.Errorf("SIAKAD category name = %q, want 'SIAKAD'", names[CategorySIAKAD])
	}
	if names[CategoryGuestPassword] != "Lupa Password SSO" {
		t.Errorf("GuestPassword category name = %q, want 'Lupa Password SSO'", names[CategoryGuestPassword])
	}
}

// ============================================================================
// Tests for Category Constants
// ============================================================================

func TestCategoryConstants(t *testing.T) {
	// Verify constants have expected values
	if CategoryInternet != "internet" {
		t.Errorf("CategoryInternet = %q, want 'internet'", CategoryInternet)
	}
	if CategorySIAKAD != "siakad" {
		t.Errorf("CategorySIAKAD = %q, want 'siakad'", CategorySIAKAD)
	}
	if CategoryWebsite != "website" {
		t.Errorf("CategoryWebsite = %q, want 'website'", CategoryWebsite)
	}
	if CategorySistemInformasi != "sistem-informasi" {
		t.Errorf("CategorySistemInformasi = %q, want 'sistem-informasi'", CategorySistemInformasi)
	}
	if CategoryLainnya != "lainnya" {
		t.Errorf("CategoryLainnya = %q, want 'lainnya'", CategoryLainnya)
	}
	if CategoryGuestPassword != "guest-password" {
		t.Errorf("CategoryGuestPassword = %q, want 'guest-password'", CategoryGuestPassword)
	}
	if CategoryGuestSSORegistration != "guest-sso" {
		t.Errorf("CategoryGuestSSORegistration = %q, want 'guest-sso'", CategoryGuestSSORegistration)
	}
	if CategoryGuestEmailRegistration != "guest-email" {
		t.Errorf("CategoryGuestEmailRegistration = %q, want 'guest-email'", CategoryGuestEmailRegistration)
	}
}

// ============================================================================
// Tests for Category filter helper function
// ============================================================================

func TestGuestCategories(t *testing.T) {
	categories := DefaultCategories()

	var guestCategories []domain.ServiceCategory
	for _, cat := range categories {
		if cat.GuestAllowed {
			guestCategories = append(guestCategories, cat)
		}
	}

	if len(guestCategories) != 3 {
		t.Errorf("Expected 3 guest-allowed categories, got %d", len(guestCategories))
	}
}

func TestRegisteredOnlyCategories(t *testing.T) {
	categories := DefaultCategories()

	var registeredCategories []domain.ServiceCategory
	for _, cat := range categories {
		if !cat.GuestAllowed {
			registeredCategories = append(registeredCategories, cat)
		}
	}

	if len(registeredCategories) != 5 {
		t.Errorf("Expected 5 registered-only categories, got %d", len(registeredCategories))
	}
}
