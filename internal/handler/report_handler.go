package handler

import (
    "encoding/csv"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"
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
	admin.GET("/reports/satisfaction/export", handler.surveySatisfactionExport)
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
	period, periods := parsePeriodParams(c, 6)
    rows, err := handler.reports.ServiceSatisfactionSummary(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) cohortReport(c *gin.Context) {
	period, periods := parsePeriodParams(c, 5)
    rows, err := handler.reports.CohortReport(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) surveySatisfaction(c *gin.Context) {
	period, periods := parsePeriodParams(c, 5)
    categoryID := c.Query("categoryId")
    templateID := c.Query("templateId")
    report, err := handler.reports.SurveySatisfaction(categoryID, templateID, period, periods)
    if err != nil {
        respondError(c, http.StatusBadRequest, err.Error())
        return
    }
	respondOK(c, report)
}

func (handler *ReportHandler) surveySatisfactionExport(c *gin.Context) {
	period, periods := parsePeriodParams(c, 5)
	categoryID := c.Query("categoryId")
	templateID := c.Query("templateId")
	report, err := handler.reports.SurveySatisfactionExport(categoryID, templateID, period, periods)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	filename := fmt.Sprintf(
		"survey_export_%s_%s_%s.csv",
		sanitizeFilename(report.CategoryID),
		sanitizeFilename(report.TemplateID),
		time.Now().Format("20060102_150405"),
	)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv; charset=utf-8")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	header := []string{
		"Kategori",
		"Template",
		"Ticket ID",
		"User ID",
		"Tanggal",
		"Skor(0-100)",
	}
	for idx, question := range report.Questions {
		header = append(header, fmt.Sprintf("Q%d - %s", idx+1, question.Text))
	}
	if err := writer.Write(header); err != nil {
		return
	}

	for _, response := range report.Responses {
		values := make([]string, 0, len(header))
		values = append(values,
			report.Category,
			report.Template,
			response.TicketID,
			response.UserID,
			response.CreatedAt.Format(time.RFC3339),
			fmt.Sprintf("%.2f", response.Score),
		)
		answers := make(map[string]interface{})
		if err := json.Unmarshal(response.Answers, &answers); err != nil {
			answers = map[string]interface{}{}
		}
		for _, question := range report.Questions {
			values = append(values, formatAnswerValue(answers[question.ID]))
		}
		if err := writer.Write(values); err != nil {
			return
		}
	}
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
	period, periods := parsePeriodParams(c, 5)
    rows, err := handler.reports.UsageCohort(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
    respondOK(c, rows)
}

func (handler *ReportHandler) entityService(c *gin.Context) {
	period, periods := parsePeriodParams(c, 5)
    rows, err := handler.reports.EntityServiceMatrix(period, periods)
    if err != nil {
        respondError(c, http.StatusInternalServerError, err.Error())
        return
    }
	respondOK(c, rows)
}

func parsePeriodParams(c *gin.Context, defaultPeriods int) (string, int) {
	periods := defaultPeriods
	if raw := c.Query("periods"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			periods = parsed
		}
	} else if raw := c.Query("months"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			periods = parsed
		}
	}
	return c.DefaultQuery("period", "monthly"), periods
}

func sanitizeFilename(value string) string {
	if value == "" {
		return "all"
	}
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_")
	return replacer.Replace(value)
}

func formatAnswerValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case bool:
		if v {
			return "Ya"
		}
		return "Tidak"
	case float64:
		if v == float64(int(v)) {
			return strconv.Itoa(int(v))
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
