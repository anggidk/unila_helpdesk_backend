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
    admin.GET("/reports", handler.serviceTrends)
    admin.GET("/reports/cohort", handler.cohortReport)
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

func (handler *ReportHandler) cohortReport(c *gin.Context) {
    months := 5
    if raw := c.Query("months"); raw != "" {
        if parsed, err := strconv.Atoi(raw); err == nil {
            months = parsed
        }
    }
    rows, err := handler.reports.CohortReport(months)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}
