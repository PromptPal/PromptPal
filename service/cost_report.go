package service

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/costreport"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	"github.com/PromptPal/PromptPal/ent/provider"
	"github.com/go-redis/cache/v9"
	"github.com/sirupsen/logrus"
)

// CostReportService handles cost report generation and caching
type CostReportService struct{}

// CostBreakdownItem represents a single item in cost breakdown
type CostBreakdownItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	CostCents float64 `json:"costCents"`
	Count     int     `json:"count"`
}

// DailyCost represents cost data for a specific day
type DailyCost struct {
	Date      string  `json:"date"`
	CostCents float64 `json:"costCents"`
	Count     int     `json:"count"`
}

// MonthlyCostReport represents aggregated cost data for a month
type MonthlyCostReport struct {
	Month   string `json:"month"`   // YYYY-MM format
	UserID  string `json:"userId"`

	// Total costs and counts
	TotalCostCents  float64 `json:"totalCostCents"`
	TotalCalls      int     `json:"totalCalls"`
	TotalTokens     int     `json:"totalTokens"`
	SuccessfulCalls int     `json:"successfulCalls"`
	CachedCalls     int     `json:"cachedCalls"`

	// Cost breakdowns by different dimensions
	CostsByProvider []CostBreakdownItem `json:"costsByProvider"`
	CostsByProject  []CostBreakdownItem `json:"costsByProject"`
	CostsByPrompt   []CostBreakdownItem `json:"costsByPrompt"`
	CostsByDay      []DailyCost         `json:"costsByDay"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetCostReports retrieves cost reports for a user within a date range
func (s *CostReportService) GetCostReports(ctx context.Context, userID, fromMonth, toMonth string) ([]MonthlyCostReport, error) {
	fromTime, err := time.Parse("2006-01", fromMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid from month format: %w", err)
	}

	toTime, err := time.Parse("2006-01", toMonth)
	if err != nil {
		return nil, fmt.Errorf("invalid to month format: %w", err)
	}

	var reports []MonthlyCostReport
	currentMonth := fromTime

	for currentMonth.Before(toTime.AddDate(0, 1, 0)) {
		monthStr := currentMonth.Format("2006-01")
		
		// Check if this is the current month (use real-time data)
		now := time.Now()
		isCurrentMonth := monthStr == now.Format("2006-01")

		var report *MonthlyCostReport
		
		if isCurrentMonth {
			// Generate real-time report for current month
			report, err = s.generateRealtimeReport(ctx, userID, monthStr)
		} else {
			// Try to get cached report first
			report, err = s.getCachedReport(ctx, userID, monthStr)
			if err != nil || report == nil {
				// Generate and cache the report
				report, err = s.generateAndCacheReport(ctx, userID, monthStr)
			}
		}

		if err != nil {
			logrus.WithError(err).Errorf("Failed to get cost report for %s/%s", userID, monthStr)
			continue
		}

		if report != nil {
			reports = append(reports, *report)
		}

		currentMonth = currentMonth.AddDate(0, 1, 0)
	}

	return reports, nil
}

// generateRealtimeReport generates a cost report in real-time for the current month
func (s *CostReportService) generateRealtimeReport(ctx context.Context, userID, month string) (*MonthlyCostReport, error) {
	monthTime, _ := time.Parse("2006-01", month)
	startOfMonth := monthTime
	endOfMonth := monthTime.AddDate(0, 1, 0)

	return s.aggregatePromptCallsForMonth(ctx, userID, month, startOfMonth, endOfMonth)
}

// generateAndCacheReport generates a cost report and caches it for historical data
func (s *CostReportService) generateAndCacheReport(ctx context.Context, userID, month string) (*MonthlyCostReport, error) {
	monthTime, _ := time.Parse("2006-01", month)
	startOfMonth := monthTime
	endOfMonth := monthTime.AddDate(0, 1, 0)

	report, err := s.aggregatePromptCallsForMonth(ctx, userID, month, startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	// Cache the report for 24 hours
	cacheKey := fmt.Sprintf("cost-report:%s:%s", userID, month)
	Cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   cacheKey,
		Value: report,
		TTL:   time.Hour * 24,
	})

	// Also save to database for persistence
	err = s.saveToDB(ctx, report)
	if err != nil {
		logrus.WithError(err).Warn("Failed to save cost report to database")
	}

	return report, nil
}

// getCachedReport retrieves a cached cost report
func (s *CostReportService) getCachedReport(ctx context.Context, userID, month string) (*MonthlyCostReport, error) {
	// Try Redis cache first
	cacheKey := fmt.Sprintf("cost-report:%s:%s", userID, month)
	var report MonthlyCostReport
	err := Cache.Get(ctx, cacheKey, &report)
	if err == nil {
		return &report, nil
	}

	// Try database if cache miss
	dbReport, err := EntClient.CostReport.Query().
		Where(costreport.UserId(userID)).
		Where(costreport.Month(month)).
		First(ctx)
	
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil // Not found, will generate new report
		}
		return nil, err
	}

	// Convert database report to service report
	return s.convertDBToServiceReport(dbReport), nil
}

// aggregatePromptCallsForMonth performs the actual aggregation of PromptCall data
func (s *CostReportService) aggregatePromptCallsForMonth(ctx context.Context, userID, month string, startTime, endTime time.Time) (*MonthlyCostReport, error) {
	report := &MonthlyCostReport{
		Month:     month,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Get total statistics
	err := s.aggregateTotalStats(ctx, userID, startTime, endTime, report)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate total stats: %w", err)
	}

	// Get costs by provider
	err = s.aggregateCostsByProvider(ctx, userID, startTime, endTime, report)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate costs by provider: %w", err)
	}

	// Get costs by project
	err = s.aggregateCostsByProject(ctx, userID, startTime, endTime, report)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate costs by project: %w", err)
	}

	// Get costs by prompt
	err = s.aggregateCostsByPrompt(ctx, userID, startTime, endTime, report)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate costs by prompt: %w", err)
	}

	// Get costs by day
	err = s.aggregateCostsByDay(ctx, userID, startTime, endTime, report)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate costs by day: %w", err)
	}

	return report, nil
}

// aggregateTotalStats aggregates total statistics for the month
func (s *CostReportService) aggregateTotalStats(ctx context.Context, userID string, startTime, endTime time.Time, report *MonthlyCostReport) error {
	baseQuery := EntClient.PromptCall.Query().
		Where(promptcall.UserId(userID)).
		Where(promptcall.CreateTimeGTE(startTime)).
		Where(promptcall.CreateTimeLT(endTime))

	// Get total count
	totalCalls, err := baseQuery.Clone().Count(ctx)
	if err != nil {
		return err
	}
	report.TotalCalls = totalCalls

	// Get successful calls count
	successfulCalls, err := baseQuery.Clone().
		Where(promptcall.Result(0)).
		Count(ctx)
	if err != nil {
		return err
	}
	report.SuccessfulCalls = successfulCalls

	// Get cached calls count
	cachedCalls, err := baseQuery.Clone().
		Where(promptcall.Cached(true)).
		Count(ctx)
	if err != nil {
		return err
	}
	report.CachedCalls = cachedCalls

	// Get aggregated cost and tokens using custom SQL
	var result []struct {
		TotalCost   float64 `json:"total_cost"`
		TotalTokens int     `json:"total_tokens"`
	}

	err = baseQuery.Clone().
		Aggregate(func(s *sql.Selector) string {
			return sql.As("COALESCE(SUM(cost_cents), 0)", "total_cost")
		}, func(s *sql.Selector) string {
			return sql.As("COALESCE(SUM(total_token), 0)", "total_tokens")
		}).
		Scan(ctx, &result)

	if err != nil {
		return err
	}

	if len(result) > 0 {
		report.TotalCostCents = result[0].TotalCost
		report.TotalTokens = result[0].TotalTokens
	}

	return nil
}

// aggregateCostsByProvider aggregates costs grouped by provider
func (s *CostReportService) aggregateCostsByProvider(ctx context.Context, userID string, startTime, endTime time.Time, report *MonthlyCostReport) error {
	var result []struct {
		ProviderID int     `json:"provider_id"`
		TotalCost  float64 `json:"total_cost"`
		Count      int     `json:"count"`
	}

	err := EntClient.PromptCall.Query().
		Where(promptcall.UserId(userID)).
		Where(promptcall.CreateTimeGTE(startTime)).
		Where(promptcall.CreateTimeLT(endTime)).
		Where(promptcall.ProviderIdNotNil()).
		GroupBy(promptcall.FieldProviderId).
		Aggregate(func(s *sql.Selector) string {
			return sql.As("COALESCE(SUM(cost_cents), 0)", "total_cost")
		}, func(s *sql.Selector) string {
			return sql.As("COUNT(*)", "count") 
		}).
		Scan(ctx, &result)

	if err != nil {
		return err
	}

	// Get provider names
	providerIDs := make([]int, len(result))
	for i, r := range result {
		providerIDs[i] = r.ProviderID
	}

	providers, err := EntClient.Provider.Query().
		Where(provider.IDIn(providerIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	providerMap := make(map[int]string)
	for _, p := range providers {
		providerMap[p.ID] = p.Name
	}

	for _, r := range result {
		name := providerMap[r.ProviderID]
		if name == "" {
			name = fmt.Sprintf("Provider %d", r.ProviderID)
		}

		report.CostsByProvider = append(report.CostsByProvider, CostBreakdownItem{
			ID:        fmt.Sprintf("%d", r.ProviderID),
			Name:      name,
			CostCents: r.TotalCost,
			Count:     r.Count,
		})
	}

	return nil
}

// aggregateCostsByProject aggregates costs grouped by project
func (s *CostReportService) aggregateCostsByProject(ctx context.Context, userID string, startTime, endTime time.Time, report *MonthlyCostReport) error {
	var result []struct {
		ProjectID int     `json:"project_id"`
		TotalCost float64 `json:"total_cost"`
		Count     int     `json:"count"`
	}

	err := EntClient.PromptCall.Query().
		Where(promptcall.UserId(userID)).
		Where(promptcall.CreateTimeGTE(startTime)).
		Where(promptcall.CreateTimeLT(endTime)).
		GroupBy("project_calls").
		Aggregate(func(s *sql.Selector) string {
			return sql.As("COALESCE(SUM(cost_cents), 0)", "total_cost")
		}, func(s *sql.Selector) string {
			return sql.As("COUNT(*)", "count") 
		}).
		Scan(ctx, &result)

	if err != nil {
		return err
	}

	// Get project names
	projectIDs := make([]int, len(result))
	for i, r := range result {
		projectIDs[i] = r.ProjectID
	}

	projects, err := EntClient.Project.Query().
		Where(project.IDIn(projectIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	projectMap := make(map[int]string)
	for _, p := range projects {
		projectMap[p.ID] = p.Name
	}

	for _, r := range result {
		name := projectMap[r.ProjectID]
		if name == "" {
			name = fmt.Sprintf("Project %d", r.ProjectID)
		}

		report.CostsByProject = append(report.CostsByProject, CostBreakdownItem{
			ID:        fmt.Sprintf("%d", r.ProjectID),
			Name:      name,
			CostCents: r.TotalCost,
			Count:     r.Count,
		})
	}

	return nil
}

// aggregateCostsByPrompt aggregates costs grouped by prompt
func (s *CostReportService) aggregateCostsByPrompt(ctx context.Context, userID string, startTime, endTime time.Time, report *MonthlyCostReport) error {
	var result []struct {
		PromptID  int     `json:"prompt_id"`
		TotalCost float64 `json:"total_cost"`
		Count     int     `json:"count"`
	}

	err := EntClient.PromptCall.Query().
		Where(promptcall.UserId(userID)).
		Where(promptcall.CreateTimeGTE(startTime)).
		Where(promptcall.CreateTimeLT(endTime)).
		GroupBy(promptcall.FieldPromptId).
		Aggregate(func(s *sql.Selector) string {
			return sql.As("COALESCE(SUM(cost_cents), 0)", "total_cost")
		}, func(s *sql.Selector) string {
			return sql.As("COUNT(*)", "count") 
		}).
		Scan(ctx, &result)

	if err != nil {
		return err
	}

	// Get prompt titles
	promptIDs := make([]int, len(result))
	for i, r := range result {
		promptIDs[i] = r.PromptID
	}

	prompts, err := EntClient.Prompt.Query().
		Where(prompt.IDIn(promptIDs...)).
		All(ctx)
	if err != nil {
		return err
	}

	promptMap := make(map[int]string)
	for _, p := range prompts {
		promptMap[p.ID] = p.Name
	}

	for _, r := range result {
		name := promptMap[r.PromptID]
		if name == "" {
			name = fmt.Sprintf("Prompt %d", r.PromptID)
		}

		report.CostsByPrompt = append(report.CostsByPrompt, CostBreakdownItem{
			ID:        fmt.Sprintf("%d", r.PromptID),
			Name:      name,
			CostCents: r.TotalCost,
			Count:     r.Count,
		})
	}

	return nil
}

// aggregateCostsByDay aggregates costs grouped by day
func (s *CostReportService) aggregateCostsByDay(ctx context.Context, userID string, startTime, endTime time.Time, report *MonthlyCostReport) error {
	var result []struct {
		Date      string  `json:"date"`
		TotalCost float64 `json:"total_cost"`
		Count     int     `json:"count"`
	}

	err := EntClient.PromptCall.Query().
		Where(promptcall.UserId(userID)).
		Where(promptcall.CreateTimeGTE(startTime)).
		Where(promptcall.CreateTimeLT(endTime)).
		Aggregate(
			func(s *sql.Selector) string {
				return sql.As("DATE(create_time)", "date")
			},
			func(s *sql.Selector) string {
				return sql.As("COALESCE(SUM(cost_cents), 0)", "total_cost")
			},
			func(s *sql.Selector) string {
				return sql.As("COUNT(*)", "count")
			},
		).
		Modify(func(s *sql.Selector) {
			s.GroupBy("date")
			s.OrderBy(sql.Asc("date"))
		}).
		Scan(ctx, &result)

	if err != nil {
		return err
	}

	for _, r := range result {
		report.CostsByDay = append(report.CostsByDay, DailyCost{
			Date:      r.Date,
			CostCents: r.TotalCost,
			Count:     r.Count,
		})
	}

	return nil
}

// saveToDB saves the cost report to the database
func (s *CostReportService) saveToDB(ctx context.Context, report *MonthlyCostReport) error {
	// Convert cost breakdowns to map[string]float64 for JSON storage
	costsByProvider := convertToStringFloatMap(report.CostsByProvider)
	costsByProject := convertToStringFloatMap(report.CostsByProject)
	costsByPrompt := convertToStringFloatMap(report.CostsByPrompt)
	costsByDay := convertDailyCostsToStringFloatMap(report.CostsByDay)

	// Create or update the cost report
	err := EntClient.CostReport.Create().
		SetMonth(report.Month).
		SetUserId(report.UserID).
		SetTotalCostCents(report.TotalCostCents).
		SetCostsByProvider(costsByProvider).
		SetCostsByProject(costsByProject).
		SetCostsByPrompt(costsByPrompt).
		SetCostsByDay(costsByDay).
		SetTotalCalls(report.TotalCalls).
		SetTotalTokens(report.TotalTokens).
		SetSuccessfulCalls(report.SuccessfulCalls).
		SetCachedCalls(report.CachedCalls).
		OnConflict().
		UpdateNewValues().
		Exec(ctx)

	return err
}

// convertDBToServiceReport converts database entity to service model
func (s *CostReportService) convertDBToServiceReport(dbReport *ent.CostReport) *MonthlyCostReport {
	report := &MonthlyCostReport{
		Month:           dbReport.Month,
		UserID:          dbReport.UserId,
		TotalCostCents:  dbReport.TotalCostCents,
		TotalCalls:      dbReport.TotalCalls,
		TotalTokens:     dbReport.TotalTokens,
		SuccessfulCalls: dbReport.SuccessfulCalls,
		CachedCalls:     dbReport.CachedCalls,
		CreatedAt:       dbReport.CreateTime,
		UpdatedAt:       dbReport.UpdateTime,
	}

	// Convert JSON fields back to structured data
	report.CostsByProvider = convertStringFloatMapToBreakdown(dbReport.CostsByProvider)
	report.CostsByProject = convertStringFloatMapToBreakdown(dbReport.CostsByProject)
	report.CostsByPrompt = convertStringFloatMapToBreakdown(dbReport.CostsByPrompt)
	report.CostsByDay = convertStringFloatMapToDailyCosts(dbReport.CostsByDay)

	return report
}

// Helper functions for JSON conversion
func convertToStringFloatMap(items []CostBreakdownItem) map[string]float64 {
	result := make(map[string]float64)
	for _, item := range items {
		result[item.ID] = item.CostCents
	}
	return result
}

func convertDailyCostsToStringFloatMap(costs []DailyCost) map[string]float64 {
	result := make(map[string]float64)
	for _, cost := range costs {
		result[cost.Date] = cost.CostCents
	}
	return result
}

func convertStringFloatMapToBreakdown(jsonData map[string]float64) []CostBreakdownItem {
	var result []CostBreakdownItem
	for id, cost := range jsonData {
		result = append(result, CostBreakdownItem{
			ID:        id,
			Name:      id, // Will be resolved later with actual names
			CostCents: cost,
			Count:     0, // Not stored in DB, would need separate aggregation
		})
	}
	return result
}

func convertStringFloatMapToDailyCosts(jsonData map[string]float64) []DailyCost {
	var result []DailyCost
	for date, cost := range jsonData {
		result = append(result, DailyCost{
			Date:      date,
			CostCents: cost,
			Count:     0, // Not stored in DB, would need separate aggregation
		})
	}
	return result
}