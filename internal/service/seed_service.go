package service

import (
	"unila_helpdesk_backend/internal/domain"
)

const (
	CategoryInternet               = "CAT001"
	CategorySIAKAD                 = "CAT002"
	CategoryWebsite                = "CAT003"
	CategorySistemInformasi        = "CAT004"
	CategoryLainnya                = "CAT005"
	CategoryGuestPassword          = "GST001"
	CategoryGuestSSORegistration   = "GST002"
	CategoryGuestEmailRegistration = "GST003"
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
