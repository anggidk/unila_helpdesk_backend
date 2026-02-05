package handler

import (
    "net/http"
    "strconv"
    "time"

    "unila_helpdesk_backend/internal/service"

    "github.com/gin-gonic/gin"
)

type ReportHandler struct {
    reports *service.ReportService
}

func NewReportHandler(reports *service.ReportService) *ReportHandler {
    return &ReportHandler{reports: reports}
}

func (handler *ReportHandler) RegisterRoutes(admin *gin.RouterGroup) {
    admin.GET("/reports/summary", handler.dashboardSummary)
    admin.GET("/reports", handler.serviceTrends)
    admin.GET("/reports/satisfaction-summary", handler.satisfactionSummary)
	admin.GET("/reports/cohort", handler.cohortReport)
	admin.GET("/reports/satisfaction", handler.surveySatisfaction)
	admin.GET("/reports/templates", handler.templatesByCategory)
	admin.GET("/reports/usage", handler.usageCohort)
	admin.GET("/reports/entity-service", handler.entityService)
}

func (handler *ReportHandler) dashboardSummary(c *gin.Context) {
    summary, err := handler.reports.DashboardSummary()
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, summary)
}

func (handler *ReportHandler) serviceTrends(c *gin.Context) {
    start := time.Now().AddDate(0, 0, -30)
    end := time.Now().AddDate(0, 0, 1)
    if rawStart := c.Query("start"); rawStart != "" {
        if parsed, err := time.Parse(time.RFC3339, rawStart); err == nil {
            start = parsed
        }
    }
    if rawEnd := c.Query("end"); rawEnd != "" {
        if parsed, err := time.Parse(time.RFC3339, rawEnd); err == nil {
            end = parsed
        }
    }
    trends, err := handler.reports.ServiceTrends(start, end)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, trends)
}

func (handler *ReportHandler) satisfactionSummary(c *gin.Context) {
    periods := 6
    if raw := c.Query("periods"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    } else if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    }
    period := c.DefaultQuery("period", "monthly")
    rows, err := handler.reports.ServiceSatisfactionSummary(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) cohortReport(c *gin.Context) {
    periods := 5
    if raw := c.Query("periods"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    } else if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    }
    period := c.DefaultQuery("period", "monthly")
    rows, err := handler.reports.CohortReport(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) surveySatisfaction(c *gin.Context) {
    periods := 5
    if raw := c.Query("periods"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    } else if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    }
    period := c.DefaultQuery("period", "monthly")
    categoryID := c.Query("categoryId")
    templateID := c.Query("templateId")
    report, err := handler.reports.SurveySatisfaction(categoryID, templateID, period, periods)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, report)
}

func (handler *ReportHandler) templatesByCategory(c *gin.Context) {
    categoryID := c.Query("categoryId")
    templates, err := handler.reports.TemplatesByCategory(categoryID)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
    respondOK(c, templates)
}

func (handler *ReportHandler) usageCohort(c *gin.Context) {
    periods := 5
    if raw := c.Query("periods"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    } else if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    }
    period := c.DefaultQuery("period", "monthly")
    rows, err := handler.reports.UsageCohort(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) entityService(c *gin.Context) {
    periods := 5
    if raw := c.Query("periods"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    } else if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            periods = parsed
        }
    }
    period := c.DefaultQuery("period", "monthly")
    rows, err := handler.reports.EntityServiceMatrix(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}
