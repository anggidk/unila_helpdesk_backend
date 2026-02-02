package handler

import (
    "net/http"

    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/repository"

    "github.com/gin-gonic/gin"
)

type CategoryHandler struct {
    categories *repository.CategoryRepository
}

var hiddenCategoryIDs = map[string]struct{}{
    "membership":     {},
    "guest-account":  {},
    "email":          {},
    "vclass":         {},
}

func NewCategoryHandler(categories *repository.CategoryRepository) *CategoryHandler {
    return &CategoryHandler{categories: categories}
}

func (handler *CategoryHandler) RegisterRoutes(public *gin.RouterGroup) {
    public.GET("/categories", handler.listAll)
    public.GET("/categories/guest", handler.listGuest)
}

func (handler *CategoryHandler) listAll(c *gin.Context) {
    items, err := handler.categories.List()
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    filtered := make([]domain.ServiceCategory, 0, len(items))
    for _, item := range items {
        if isHiddenCategory(item.ID) {
            continue
        }
        filtered = append(filtered, item)
    }
    respondOK(c, toCategoryDTOs(filtered))
}

func (handler *CategoryHandler) listGuest(c *gin.Context) {
    items, err := handler.categories.List()
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    filtered := make([]domain.ServiceCategory, 0)
    for _, item := range items {
        if isHiddenCategory(item.ID) {
            continue
        }
        if item.GuestAllowed {
            filtered = append(filtered, item)
        }
    }
    respondOK(c, toCategoryDTOs(filtered))
}

func isHiddenCategory(id string) bool {
    _, hidden := hiddenCategoryIDs[id]
    return hidden
}

func toCategoryDTOs(items []domain.ServiceCategory) []domain.ServiceCategoryDTO {
    result := make([]domain.ServiceCategoryDTO, 0, len(items))
    for _, item := range items {
        result = append(result, domain.ServiceCategoryDTO{
            ID:           item.ID,
            Name:         item.Name,
            GuestAllowed: item.GuestAllowed,
        })
    }
    return result
}
