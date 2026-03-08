package service

import (
	"unila_helpdesk_backend/internal/domain"
)

const (
	CategoryGuestPassword        = domain.ServiceGuestPassword
	CategoryGuestSSORegistration = domain.ServiceGuestRegistration
	CategoryGuestEmail           = domain.ServiceGuestEmail
	CategoryInternet             = domain.ServiceInternet
	CategoryWebsite              = domain.ServiceWebsiteDown
	CategorySistemInformasi      = domain.ServiceSistemInformasi
	CategorySIAKAD               = domain.ServiceSIAKADU
	CategoryLainnya              = domain.ServiceLainnya
)

func DefaultCategories() []domain.ServiceCategory {
	return []domain.ServiceCategory{
		{ID: CategoryGuestPassword, Name: "Lupa Password SSO", GuestAllowed: true},
		{ID: CategoryGuestSSORegistration, Name: "Registrasi SSO", GuestAllowed: true},
		{ID: CategoryGuestEmail, Name: "Email Resmi Unila", GuestAllowed: true},
		{ID: CategoryInternet, Name: "Jaringan Internet", GuestAllowed: false},
		{ID: CategoryWebsite, Name: "Website Down", GuestAllowed: false},
		{ID: CategorySistemInformasi, Name: "Sistem Informasi", GuestAllowed: false},
		{ID: CategorySIAKAD, Name: "SIAKADU", GuestAllowed: false},
		{ID: CategoryLainnya, Name: "Lainnya", GuestAllowed: false},
	}
}
