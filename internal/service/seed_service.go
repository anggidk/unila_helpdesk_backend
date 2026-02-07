package service

import (
	"fmt"

	"unila_helpdesk_backend/internal/domain"

	"gorm.io/gorm"
)

const (
	CategoryInternet            = "internet"
	CategorySIAKAD              = "siakad"
	CategoryWebsite             = "website"
	CategorySistemInformasi     = "sistem-informasi"
	CategoryLainnya             = "lainnya"
	CategoryGuestPassword       = "guest-password"
	CategoryGuestSSORegistration = "guest-sso"
	CategoryGuestEmailRegistration = "guest-email-unila"
)

func DefaultCategories() []domain.ServiceCategory {
	return []domain.ServiceCategory{
		{ID: CategoryInternet, Name: "Jaringan Internet", GuestAllowed: false},
		{ID: CategorySIAKAD, Name: "SIAKAD", GuestAllowed: false},
		{ID: CategoryWebsite, Name: "Website", GuestAllowed: false},
		{ID: CategorySistemInformasi, Name: "Sistem Informasi", GuestAllowed: false},
		{ID: CategoryLainnya, Name: "Lainnya", GuestAllowed: false},
		{ID: CategoryGuestPassword, Name: "Lupa Password SSO", GuestAllowed: true},
		{ID: CategoryGuestSSORegistration, Name: "Registrasi SSO", GuestAllowed: true},
		{ID: CategoryGuestEmailRegistration, Name: "Registrasi Email @unila.ac.id", GuestAllowed: true},
	}
}

func CleanupDeprecatedCategories(database *gorm.DB, fallbackCategoryID string) error {
	if fallbackCategoryID == "" {
		return fmt.Errorf("fallback category id wajib diisi")
	}
	deprecatedIDs := []string{
		"membership",
		"guest-account",
		"guest-email",
		"email",
		"vclass",
	}
	return database.Transaction(func(tx *gorm.DB) error {
		var fallbackCount int64
		if err := tx.Model(&domain.ServiceCategory{}).
			Where("id = ?", fallbackCategoryID).
			Count(&fallbackCount).Error; err != nil {
			return err
		}
		if fallbackCount == 0 {
			return fmt.Errorf("fallback category '%s' tidak ditemukan", fallbackCategoryID)
		}

		if err := tx.Model(&domain.Ticket{}).
			Where("category_id IN ?", deprecatedIDs).
			Update("category_id", fallbackCategoryID).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.SurveyTemplate{}).
			Where("category_id IN ?", deprecatedIDs).
			Update("category_id", fallbackCategoryID).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", deprecatedIDs).
			Delete(&domain.ServiceCategory{}).Error; err != nil {
			return err
		}
		return nil
	})
}
