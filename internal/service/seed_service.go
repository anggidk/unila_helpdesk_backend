package service

import "unila_helpdesk_backend/internal/domain"

const (
    CategoryInternet   = "internet"
    CategoryWebsite    = "website"
    CategoryVClass     = "vclass"
    CategorySIAKAD     = "siakad"
    CategoryEmail      = "email"
    CategoryMembership = "membership"
    CategoryGuestPassword = "guest-password"
    CategoryGuestAccount  = "guest-account"
    CategoryGuestEmail    = "guest-email"
)

func DefaultCategories() []domain.ServiceCategory {
    return []domain.ServiceCategory{
        {ID: CategoryInternet, Name: "Jaringan Internet", GuestAllowed: false},
        {ID: CategoryWebsite, Name: "Website Layanan", GuestAllowed: false},
        {ID: CategoryVClass, Name: "VClass", GuestAllowed: false},
        {ID: CategorySIAKAD, Name: "SIAKAD", GuestAllowed: false},
        {ID: CategoryEmail, Name: "Email Unila", GuestAllowed: false},
        {ID: CategoryMembership, Name: "Keanggotaan", GuestAllowed: true},
        {ID: CategoryGuestPassword, Name: "Lupa Password", GuestAllowed: true},
        {ID: CategoryGuestAccount, Name: "Buat Akun Baru", GuestAllowed: true},
        {ID: CategoryGuestEmail, Name: "Buat Email @unila.ac.id", GuestAllowed: true},
    }
}
