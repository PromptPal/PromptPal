package schema

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/service"
)

// reports(filters: CostReportFilters!): CostReportList!
type costReportFilters struct {
	From string
	To   string
}

type reportsArgs struct {
	Filters costReportFilters
}

type costReportListResponse struct {
	reports []service.MonthlyCostReport
}

// Reports resolves the main cost reports query
func (q QueryResolver) Reports(ctx context.Context, args reportsArgs) (costReportListResponse, error) {
	// Get user ID from context
	ctxValue := ctx.Value(service.GinGraphQLContextKey).(service.GinGraphQLContextType)
	userID := fmt.Sprintf("%d", ctxValue.UserID)

	// Validate date format
	if _, err := time.Parse("2006-01", args.Filters.From); err != nil {
		return costReportListResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
	}
	if _, err := time.Parse("2006-01", args.Filters.To); err != nil {
		return costReportListResponse{}, NewGraphQLHttpError(http.StatusBadRequest, err)
	}

	// Use the cost report service to get the reports
	costReportService := &service.CostReportService{}
	reports, err := costReportService.GetCostReports(ctx, userID, args.Filters.From, args.Filters.To)
	if err != nil {
		return costReportListResponse{}, NewGraphQLHttpError(http.StatusInternalServerError, err)
	}

	return costReportListResponse{reports: reports}, nil
}

// Reports method returns the list of monthly cost reports
func (r costReportListResponse) Reports() []monthlyCostReportResponse {
	var result []monthlyCostReportResponse
	for _, report := range r.reports {
		result = append(result, monthlyCostReportResponse{report: report})
	}
	return result
}

// Count returns the number of reports
func (r costReportListResponse) Count() int32 {
	return int32(len(r.reports))
}

// monthlyCostReportResponse wraps a MonthlyCostReport for GraphQL
type monthlyCostReportResponse struct {
	report service.MonthlyCostReport
}

func (r monthlyCostReportResponse) Month() string {
	return r.report.Month
}

func (r monthlyCostReportResponse) UserId() string {
	return r.report.UserID
}

func (r monthlyCostReportResponse) TotalCostCents() float64 {
	return r.report.TotalCostCents
}

func (r monthlyCostReportResponse) TotalCalls() int32 {
	return int32(r.report.TotalCalls)
}

func (r monthlyCostReportResponse) TotalTokens() int32 {
	return int32(r.report.TotalTokens)
}

func (r monthlyCostReportResponse) SuccessfulCalls() int32 {
	return int32(r.report.SuccessfulCalls)
}

func (r monthlyCostReportResponse) CachedCalls() int32 {
	return int32(r.report.CachedCalls)
}

func (r monthlyCostReportResponse) CostsByProvider() []costBreakdownItemResponse {
	var result []costBreakdownItemResponse
	for _, item := range r.report.CostsByProvider {
		result = append(result, costBreakdownItemResponse{item: item})
	}
	return result
}

func (r monthlyCostReportResponse) CostsByProject() []costBreakdownItemResponse {
	var result []costBreakdownItemResponse
	for _, item := range r.report.CostsByProject {
		result = append(result, costBreakdownItemResponse{item: item})
	}
	return result
}

func (r monthlyCostReportResponse) CostsByPrompt() []costBreakdownItemResponse {
	var result []costBreakdownItemResponse
	for _, item := range r.report.CostsByPrompt {
		result = append(result, costBreakdownItemResponse{item: item})
	}
	return result
}

func (r monthlyCostReportResponse) CostsByDay() []dailyCostResponse {
	var result []dailyCostResponse
	for _, cost := range r.report.CostsByDay {
		result = append(result, dailyCostResponse{cost: cost})
	}
	return result
}

func (r monthlyCostReportResponse) CreatedAt() string {
	return r.report.CreatedAt.Format(time.RFC3339)
}

func (r monthlyCostReportResponse) UpdatedAt() string {
	return r.report.UpdatedAt.Format(time.RFC3339)
}

// costBreakdownItemResponse wraps a CostBreakdownItem for GraphQL
type costBreakdownItemResponse struct {
	item service.CostBreakdownItem
}

func (r costBreakdownItemResponse) ID() string {
	return r.item.ID
}

func (r costBreakdownItemResponse) Name() string {
	return r.item.Name
}

func (r costBreakdownItemResponse) CostCents() float64 {
	return r.item.CostCents
}

func (r costBreakdownItemResponse) Count() int32 {
	return int32(r.item.Count)
}

// dailyCostResponse wraps a DailyCost for GraphQL
type dailyCostResponse struct {
	cost service.DailyCost
}

func (r dailyCostResponse) Date() string {
	return r.cost.Date
}

func (r dailyCostResponse) CostCents() float64 {
	return r.cost.CostCents
}

func (r dailyCostResponse) Count() int32 {
	return int32(r.cost.Count)
}