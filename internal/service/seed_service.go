package service

import "unila_helpdesk_backend/internal/domain"

const (
    CategoryInternet   = "internet"
    CategorySIAKAD     = "siakad"
    CategoryWebsite    = "website"
    CategorySistemInformasi = "sistem-informasi"
    CategoryLainnya    = "lainnya"
    CategoryGuestPassword = "guest-password"
    CategoryGuestSSORegistration = "guest-sso"
    CategoryGuestEmailRegistration = "guest-email"
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
