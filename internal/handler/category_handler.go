package handler

import (
	"net/http"

	"unila_helpdesk_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categories *service.CategoryService
}

func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (handler *CategoryHandler) RegisterRoutes(public *gin.RouterGroup) {
	public.GET("/categories", handler.listAll)
	public.GET("/categories/guest", handler.listGuest)
}

func (handler *CategoryHandler) RegisterAdminRoutes(admin *gin.RouterGroup) {
	admin.PUT("/categories/:id/template", handler.assignTemplate)
}

func (handler *CategoryHandler) listAll(c *gin.Context) {
	items, err := handler.categories.ListAll()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, items)
}

func (handler *CategoryHandler) listGuest(c *gin.Context) {
	items, err := handler.categories.ListGuest()
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(c, items)
}

type assignTemplateRequest struct {
	TemplateID string `json:"templateId"`
}

func (handler *CategoryHandler) assignTemplate(c *gin.Context) {
	categoryID := c.Param("id")
	var req assignTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "payload tidak valid")
		return
	}
	if err := handler.categories.AssignTemplate(categoryID, req.TemplateID); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondOK(c, gin.H{"categoryId": categoryID, "templateId": req.TemplateID})
}
