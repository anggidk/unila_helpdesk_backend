package service

import (
	"errors"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
)

type CategoryService struct {
	categories *repository.CategoryRepository
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (service *CategoryService) ListAll() ([]domain.ServiceCategoryDTO, error) {
	items, err := service.categories.List()
	if err != nil {
		return nil, err
	}
	return toCategoryDTOs(items), nil
}

func (service *CategoryService) ListGuest() ([]domain.ServiceCategoryDTO, error) {
	items, err := service.categories.List()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.ServiceCategory, 0)
	for _, item := range items {
		if item.GuestAllowed {
			filtered = append(filtered, item)
		}
	}
	return toCategoryDTOs(filtered), nil
}

func (service *CategoryService) AssignTemplate(categoryID string, templateID string) error {
	if categoryID == "" {
		return errors.New("kategori tidak ditemukan")
	}
	return service.categories.UpdateTemplate(categoryID, templateID)
}

func toCategoryDTOs(items []domain.ServiceCategory) []domain.ServiceCategoryDTO {
	result := make([]domain.ServiceCategoryDTO, 0, len(items))
	for _, item := range items {
		result = append(result, domain.ServiceCategoryDTO{
			ID:           item.ID,
			Name:         item.Name,
			GuestAllowed: item.GuestAllowed,
			TemplateID:   item.SurveyTemplateID,
		})
	}
	return result
}
