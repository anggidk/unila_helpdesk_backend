package handler

import (
    "net/http"
    "strconv"
    "time"

    "unila_helpdesk_backend/internal/domain"
    "unila_helpdesk_backend/internal/middleware"
    "unila_helpdesk_backend/internal/repository"
    "unila_helpdesk_backend/internal/service"

    "github.com/gin-gonic/gin"
)

type SurveyHandler struct {
    surveys *service.SurveyService
}

func NewSurveyHandler(surveys *service.SurveyService) *SurveyHandler {
    return &SurveyHandler{surveys: surveys}
}

func (handler *SurveyHandler) RegisterRoutes(public *gin.RouterGroup, auth *gin.RouterGroup, admin *gin.RouterGroup) {
    public.GET("/surveys", handler.listTemplates)
    public.GET("/surveys/categories/:categoryId", handler.templateByCategory)
    admin.POST("/surveys", handler.createTemplate)
    admin.POST("/surveys/templates", handler.createTemplate)
    admin.PUT("/surveys/templates/:id", handler.updateTemplate)
    admin.DELETE("/surveys/templates/:id", handler.deleteTemplate)
    admin.GET("/surveys/responses", handler.listResponses)
    auth.POST("/surveys/responses", handler.submitResponse)
}

func (handler *SurveyHandler) listTemplates(c *gin.Context) {
    templates, err := handler.surveys.ListTemplates()
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, templates)
}

func (handler *SurveyHandler) templateByCategory(c *gin.Context) {
    categoryID := c.Param("categoryId")
    template, err := handler.surveys.TemplateByCategory(categoryID)
    if err != nil {
        respondError(c, http.StatusNotFound, "template tidak ditemukan")
        return
    }
    respondOK(c, template)
}

func (handler *SurveyHandler) createTemplate(c *gin.Context) {
    var req service.SurveyTemplateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    template, err := handler.surveys.CreateTemplate(req)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondCreated(c, template)
}

func (handler *SurveyHandler) updateTemplate(c *gin.Context) {
    templateID := c.Param("id")
    var req service.SurveyTemplateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    template, err := handler.surveys.UpdateTemplate(templateID, req)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, template)
}

func (handler *SurveyHandler) deleteTemplate(c *gin.Context) {
    templateID := c.Param("id")
    if err := handler.surveys.DeleteTemplate(templateID); err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, gin.H{"deleted": true})
}

func (handler *SurveyHandler) submitResponse(c *gin.Context) {
    user, ok := middleware.GetUser(c)
    if !ok {
        respondError(c, http.StatusUnauthorized, "token dibutuhkan")
        return
    }
    if user.Role != domain.RoleRegistered {
        respondError(c, http.StatusForbidden, "hanya pengguna terdaftar yang dapat mengisi survey")
        return
    }
    var req service.SurveyResponseRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        respondError(c, http.StatusBadRequest, "payload tidak valid")
        return
    }
    if err := handler.surveys.SubmitSurvey(user, req); err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, gin.H{"submitted": true})
}

func (handler *SurveyHandler) listResponses(c *gin.Context) {
    query := c.Query("q")
    categoryID := c.Query("categoryId")
    templateID := c.Query("templateId")
    startRaw := c.Query("start")
    endRaw := c.Query("end")

    var start *time.Time
    if startRaw != "" {
        parsed, err := time.Parse(time.RFC3339, startRaw)
        if err != nil {
            respondError(c, http.StatusBadRequest, "start tidak valid")
            return
        }
        start = &parsed
    }
    var end *time.Time
    if endRaw != "" {
        parsed, err := time.Parse(time.RFC3339, endRaw)
        if err != nil {
            respondError(c, http.StatusBadRequest, "end tidak valid")
            return
        }
        end = &parsed
    }

    page := 1
    if raw := c.Query("page"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
            page = parsed
        }
    }
    limit := 50
    if raw := c.Query("limit"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
            limit = parsed
        }
    }

    result, err := handler.surveys.ListResponsesPaged(repository.SurveyResponseFilter{
        Query:      query,
        CategoryID: categoryID,
        TemplateID: templateID,
        Start:      start,
        End:        end,
    }, page, limit)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, result)
}
