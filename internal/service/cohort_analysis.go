package service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"unila_helpdesk_backend/internal/domain"
	"unila_helpdesk_backend/internal/repository"
)

type cohortAccumulator struct {
	Label        string
	SortTime     time.Time
	UserSet      map[string]struct{}
	EligibleSets []map[string]struct{}
	ActiveSets   []map[string]struct{}
	ScoreSums    []float64
	ScoreCounts  []int
}

func (service *ReportService) CohortReport(
	period string,
	lookback int,
	buckets int,
) (domain.CohortAnalysisDTO, error) {
	unit := normalizePeriod(period)
	lookback = normalizeCohortLookback(unit, lookback)
	buckets = normalizeCohortBuckets(unit, buckets)

	start, end := periodRange(unit, lookback, service.now)
	events, err := service.reports.ListRegisteredSurveyEvents(end)
	if err != nil {
		return domain.CohortAnalysisDTO{}, err
	}

	categories := service.categoryNameMap()
	lastBucketStart := addPeriods(end, unit, -1)
	overallAccumulators := make(map[string]*cohortAccumulator)
	serviceAccumulators := make(map[string]*cohortAccumulator)
	entityAccumulators := make(map[string]*cohortAccumulator)
	userEvents := make(map[string][]repository.RegisteredSurveyEventRow)

	for _, event := range events {
		userID := strings.TrimSpace(event.UserID)
		if userID == "" {
			continue
		}
		userEvents[userID] = append(userEvents[userID], event)
	}

	for userID, rows := range userEvents {
		if len(rows) == 0 {
			continue
		}

		first := rows[0]
		anchorStart := periodStart(first.CreatedAt, unit)
		if anchorStart.Before(start) || !anchorStart.Before(end) {
			continue
		}

		maxAvailable := periodDiff(anchorStart, lastBucketStart, unit)
		if maxAvailable < 0 {
			continue
		}
		if maxAvailable >= buckets {
			maxAvailable = buckets - 1
		}

		overallKey := anchorStart.Format(time.RFC3339)
		overallAcc := ensureCohortAccumulator(
			overallAccumulators,
			overallKey,
			formatCohortLabel(anchorStart, unit),
			anchorStart,
			buckets,
		)

		serviceKey := strconv.Itoa(first.ServiceID)
		serviceLabel := categories[serviceKey]
		if serviceLabel == "" {
			serviceLabel = serviceKey
		}
		serviceAcc := ensureCohortAccumulator(
			serviceAccumulators,
			serviceKey,
			serviceLabel,
			time.Time{},
			buckets,
		)

		entityLabel := strings.TrimSpace(first.Entity)
		if entityLabel == "" {
			entityLabel = domain.EntityLainnya
		}
		entityAcc := ensureCohortAccumulator(
			entityAccumulators,
			entityLabel,
			entityLabel,
			time.Time{},
			buckets,
		)

		accumulators := []*cohortAccumulator{overallAcc, serviceAcc, entityAcc}
		for _, acc := range accumulators {
			acc.UserSet[userID] = struct{}{}
			for age := 0; age <= maxAvailable; age++ {
				acc.EligibleSets[age][userID] = struct{}{}
			}
		}

		activeBucketSeen := make(map[int]struct{})
		for _, row := range rows {
			age := periodDiff(anchorStart, periodStart(row.CreatedAt, unit), unit)
			if age < 0 || age >= buckets {
				continue
			}

			if _, seen := activeBucketSeen[age]; !seen {
				for _, acc := range accumulators {
					acc.ActiveSets[age][userID] = struct{}{}
				}
				activeBucketSeen[age] = struct{}{}
			}

			score := scoreToFivePoint(row.Score)
			if score <= 0 {
				continue
			}
			for _, acc := range accumulators {
				acc.ScoreSums[age] += score
				acc.ScoreCounts[age]++
			}
		}
	}

	report := domain.CohortAnalysisDTO{
		Period:             unit,
		LookbackPeriods:    lookback,
		BucketCount:        buckets,
		Start:              start,
		End:                end,
		BucketLabels:       buildCohortBucketLabels(unit, buckets),
		Overall:            cohortRowsFromAccumulators(overallAccumulators, buckets, true),
		ServiceComparisons: cohortRowsFromAccumulators(serviceAccumulators, buckets, false),
		EntityComparisons:  cohortRowsFromAccumulators(entityAccumulators, buckets, false),
	}
	satisfactionOverview, err := service.buildSatisfactionOverview(unit, start, end)
	if err != nil {
		return domain.CohortAnalysisDTO{}, err
	}
	report.SatisfactionOverview = &satisfactionOverview
	report.Insights = buildCohortInsights(report)
	return report, nil
}

func ensureCohortAccumulator(
	store map[string]*cohortAccumulator,
	key string,
	label string,
	sortTime time.Time,
	periods int,
) *cohortAccumulator {
	if existing, ok := store[key]; ok {
		return existing
	}

	acc := &cohortAccumulator{
		Label:        label,
		SortTime:     sortTime,
		UserSet:      make(map[string]struct{}),
		EligibleSets: make([]map[string]struct{}, periods),
		ActiveSets:   make([]map[string]struct{}, periods),
		ScoreSums:    make([]float64, periods),
		ScoreCounts:  make([]int, periods),
	}
	for index := 0; index < periods; index++ {
		acc.EligibleSets[index] = make(map[string]struct{})
		acc.ActiveSets[index] = make(map[string]struct{})
	}
	store[key] = acc
	return acc
}

func cohortRowsFromAccumulators(
	accumulators map[string]*cohortAccumulator,
	periods int,
	sortByTime bool,
) []domain.CohortAnalysisRowDTO {
	rows := make([]domain.CohortAnalysisRowDTO, 0, len(accumulators))
	for _, acc := range accumulators {
		buckets := make([]domain.CohortBucketDTO, periods)
		for age := 0; age < periods; age++ {
			eligibleCount := len(acc.EligibleSets[age])
			if eligibleCount == 0 {
				buckets[age] = domain.CohortBucketDTO{}
				continue
			}

			activeCount := len(acc.ActiveSets[age])
			retention := roundTo(float64(activeCount)/float64(eligibleCount)*100, 1)
			eligibleValue := eligibleCount
			activeValue := activeCount
			retentionValue := retention
			bucket := domain.CohortBucketDTO{
				EligibleUsers: &eligibleValue,
				ActiveUsers:   &activeValue,
				Retention:     &retentionValue,
			}
			if acc.ScoreCounts[age] > 0 {
				avgScore := roundTo(acc.ScoreSums[age]/float64(acc.ScoreCounts[age]), 2)
				bucket.AvgScore = &avgScore
			}
			buckets[age] = bucket
		}

		row := domain.CohortAnalysisRowDTO{
			Label:      acc.Label,
			Users:      len(acc.UserSet),
			Buckets:    buckets,
			DropOff:    bucketDropOff(buckets),
			ScoreDelta: bucketScoreDelta(buckets),
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		if sortByTime {
			left := accumulators[rowKeyForLabel(accumulators, rows[i].Label)]
			right := accumulators[rowKeyForLabel(accumulators, rows[j].Label)]
			if left != nil && right != nil && !left.SortTime.Equal(right.SortTime) {
				return left.SortTime.Before(right.SortTime)
			}
		}
		if rows[i].Users != rows[j].Users && !sortByTime {
			return rows[i].Users > rows[j].Users
		}
		return rows[i].Label < rows[j].Label
	})

	return rows
}

func rowKeyForLabel(
	accumulators map[string]*cohortAccumulator,
	label string,
) string {
	for key, acc := range accumulators {
		if acc.Label == label {
			return key
		}
	}
	return ""
}

func buildCohortBucketLabels(unit string, periods int) []string {
	prefix := "M"
	switch unit {
	case "daily":
		prefix = "D"
	case "weekly":
		prefix = "W"
	case "yearly":
		prefix = "Y"
	}

	labels := make([]string, 0, periods)
	for index := 0; index < periods; index++ {
		labels = append(labels, fmt.Sprintf("%s%d", prefix, index))
	}
	return labels
}

func periodDiff(start time.Time, end time.Time, unit string) int {
	start = periodStart(start, unit)
	end = periodStart(end, unit)

	switch unit {
	case "daily":
		return int(end.Sub(start).Hours() / 24)
	case "weekly":
		return int(end.Sub(start).Hours() / (24 * 7))
	case "yearly":
		return end.Year() - start.Year()
	default:
		return (end.Year()-start.Year())*12 + int(end.Month()-start.Month())
	}
}

func roundTo(value float64, decimals int) float64 {
	if decimals <= 0 {
		return math.Round(value)
	}
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

func bucketDropOff(buckets []domain.CohortBucketDTO) *float64 {
	if len(buckets) < 2 || buckets[0].Retention == nil || buckets[1].Retention == nil {
		return nil
	}
	value := roundTo(*buckets[0].Retention-*buckets[1].Retention, 1)
	return &value
}

func bucketScoreDelta(buckets []domain.CohortBucketDTO) *float64 {
	var first *float64
	var last *float64
	for _, bucket := range buckets {
		if bucket.AvgScore == nil {
			continue
		}
		if first == nil {
			value := *bucket.AvgScore
			first = &value
		}
		value := *bucket.AvgScore
		last = &value
	}
	if first == nil || last == nil {
		return nil
	}
	value := roundTo(*last-*first, 2)
	return &value
}

func buildCohortInsights(report domain.CohortAnalysisDTO) []domain.CohortInsightDTO {
	insights := make([]domain.CohortInsightDTO, 0, 4)

	if len(report.BucketLabels) > 1 {
		if bestRows, value, ok := bestRetentionRows(report.Overall, 1, true); ok {
			insights = append(insights, domain.CohortInsightDTO{
				Title: tieAwareTitle("Retensi Awal Terbaik", len(bestRows)),
				Value: fmt.Sprintf("%.1f%%", value),
				Detail: fmt.Sprintf(
					"%s %s %.1f%% pengguna pada %s.",
					formatRowLabelGroup("Cohort", bestRows),
					tieAwareVerb("mempertahankan", len(bestRows)),
					value,
					report.BucketLabels[1],
				),
			})
		}

		if weakestRows, value, ok := bestRetentionRows(report.Overall, 1, false); ok {
			insights = append(insights, domain.CohortInsightDTO{
				Title: tieAwareTitle("Retensi Awal Terendah", len(weakestRows)),
				Value: fmt.Sprintf("%.1f%%", value),
				Detail: fmt.Sprintf(
					"%s %s retensi %.1f%% pada %s.",
					formatRowLabelGroup("Cohort", weakestRows),
					tieAwareVerb("memiliki", len(weakestRows)),
					value,
					report.BucketLabels[1],
				),
			})
		}

		if serviceRows, value, ok := largestDropOffRows(report.ServiceComparisons); ok {
			insights = append(insights, domain.CohortInsightDTO{
				Title: tieAwareTitle("Drop-off Layanan Terbesar", len(serviceRows)),
				Value: fmt.Sprintf("%.1f poin", value),
				Detail: fmt.Sprintf(
					"%s %s %.1f poin dari %s ke %s.",
					formatRowLabelGroup("Kelompok layanan", serviceRows),
					tieAwareVerb("turun", len(serviceRows)),
					value,
					report.BucketLabels[0],
					report.BucketLabels[1],
				),
			})
		}
	}

	if entityRow, stability := mostStableRow(report.EntityComparisons); entityRow != nil {
		insights = append(insights, domain.CohortInsightDTO{
			Title:  "Entitas Paling Stabil",
			Value:  fmt.Sprintf("%.1f poin", stability),
			Detail: fmt.Sprintf("Entitas %s memiliki perubahan retensi rata-rata %.1f poin antar bucket.", entityRow.Label, stability),
		})
	}

	return insights
}

func bestRetentionRows(
	rows []domain.CohortAnalysisRowDTO,
	bucketIndex int,
	pickMax bool,
) ([]domain.CohortAnalysisRowDTO, float64, bool) {
	selected := make([]domain.CohortAnalysisRowDTO, 0)
	var selectedValue float64
	for index := range rows {
		value, ok := bucketRetention(rows[index], bucketIndex)
		if !ok {
			continue
		}
		if len(selected) == 0 {
			selected = append(selected, rows[index])
			selectedValue = value
			continue
		}

		if nearlyEqual(value, selectedValue) {
			selected = append(selected, rows[index])
			continue
		}

		if (pickMax && value > selectedValue) || (!pickMax && value < selectedValue) {
			selected = selected[:0]
			selected = append(selected, rows[index])
			selectedValue = value
		}
	}
	if len(selected) == 0 {
		return nil, 0, false
	}
	return selected, selectedValue, true
}

func largestDropOffRows(
	rows []domain.CohortAnalysisRowDTO,
) ([]domain.CohortAnalysisRowDTO, float64, bool) {
	selected := make([]domain.CohortAnalysisRowDTO, 0)
	var selectedValue float64
	for index := range rows {
		if rows[index].DropOff == nil {
			continue
		}
		value := *rows[index].DropOff
		if len(selected) == 0 {
			selected = append(selected, rows[index])
			selectedValue = value
			continue
		}
		if nearlyEqual(value, selectedValue) {
			selected = append(selected, rows[index])
			continue
		}
		if value > selectedValue {
			selected = selected[:0]
			selected = append(selected, rows[index])
			selectedValue = value
		}
	}
	if len(selected) == 0 {
		return nil, 0, false
	}
	return selected, selectedValue, true
}

func mostStableRow(
	rows []domain.CohortAnalysisRowDTO,
) (*domain.CohortAnalysisRowDTO, float64) {
	var (
		selected      *domain.CohortAnalysisRowDTO
		selectedValue float64
	)
	for index := range rows {
		value, ok := retentionStability(rows[index].Buckets)
		if !ok {
			continue
		}
		if selected == nil || value < selectedValue {
			selected = &rows[index]
			selectedValue = value
		}
	}
	return selected, selectedValue
}

func strongestScoreShiftRow(
	rows []domain.CohortAnalysisRowDTO,
) *domain.CohortAnalysisRowDTO {
	var selected *domain.CohortAnalysisRowDTO
	for index := range rows {
		if rows[index].ScoreDelta == nil {
			continue
		}
		if selected == nil || math.Abs(*rows[index].ScoreDelta) > math.Abs(*selected.ScoreDelta) {
			selected = &rows[index]
		}
	}
	return selected
}

func bucketRetention(row domain.CohortAnalysisRowDTO, index int) (float64, bool) {
	if index < 0 || index >= len(row.Buckets) || row.Buckets[index].Retention == nil {
		return 0, false
	}
	return *row.Buckets[index].Retention, true
}

func bucketRetentionValue(row domain.CohortAnalysisRowDTO, index int) float64 {
	value, _ := bucketRetention(row, index)
	return value
}

func retentionStability(buckets []domain.CohortBucketDTO) (float64, bool) {
	values := make([]float64, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.Retention == nil {
			continue
		}
		values = append(values, *bucket.Retention)
	}
	if len(values) < 2 {
		return 0, false
	}
	var total float64
	for index := 1; index < len(values); index++ {
		total += math.Abs(values[index] - values[index-1])
	}
	return roundTo(total/float64(len(values)-1), 1), true
}

func formatSignedFloat(value float64) string {
	if value > 0 {
		return fmt.Sprintf("+%.2f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func nearlyEqual(left float64, right float64) bool {
	return math.Abs(left-right) < 0.0001
}

func tieAwareTitle(base string, total int) string {
	if total > 1 {
		return base + " (Seri)"
	}
	return base
}

func tieAwareVerb(base string, total int) string {
	if total > 1 {
		return "sama-sama " + base
	}
	return base
}

func formatRowLabelGroup(prefix string, rows []domain.CohortAnalysisRowDTO) string {
	labels := make([]string, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, row.Label)
	}
	return formatLabelGroup(prefix, labels)
}

func formatLabelGroup(prefix string, labels []string) string {
	if len(labels) == 0 {
		return prefix
	}
	if len(labels) == 1 {
		return fmt.Sprintf("%s %s", prefix, labels[0])
	}
	if len(labels) == 2 {
		return fmt.Sprintf("%s %s dan %s", prefix, labels[0], labels[1])
	}
	if len(labels) == 3 {
		return fmt.Sprintf("%s %s, %s, dan %s", prefix, labels[0], labels[1], labels[2])
	}
	return fmt.Sprintf("%s %s, %s, dan %d lainnya", prefix, labels[0], labels[1], len(labels)-2)
}
