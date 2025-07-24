package schema

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type costReportTestSuite struct {
	suite.Suite
	uid       int
	projectID int
	promptID  int
	provider  *ent.Provider
	q         QueryResolver
	ctx       context.Context
}

func (s *costReportTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewWeb3Service()
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)

	rbac := service.NewMockRBACService(s.T())
	// Configure mock expectations for RBAC permissions
	Setup(hs, w3, rbac)

	s.q = QueryResolver{}

	// Create test user
	uniqueAddr := fmt.Sprintf("test-cost-report-%s", utils.RandStringRunes(8))
	uniqueEmail := fmt.Sprintf("test-cost-report-%s@annatarhe.com", utils.RandStringRunes(8))
	u := service.
		EntClient.
		User.
		Create().
		SetAddr(uniqueAddr).
		SetName(utils.RandStringRunes(16)).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(255).
		SetEmail(uniqueEmail).
		SaveX(context.Background())
	s.uid = u.ID

	s.ctx = context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})

	// Create test project
	project := service.
		EntClient.
		Project.
		Create().
		SetName(fmt.Sprintf("Test Project %s", utils.RandStringRunes(8))).
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.projectID = project.ID

	// Create test provider
	provider := service.
		EntClient.
		Provider.
		Create().
		SetName(fmt.Sprintf("Test Provider %s", utils.RandStringRunes(8))).
		SetEndpoint("https://api.openai.com/v1").
		SetApiKey("sk-test").
		SetSource("openai").
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.provider = provider

	// Create test prompt
	prompt := service.
		EntClient.
		Prompt.
		Create().
		SetName(fmt.Sprintf("Test Prompt %s", utils.RandStringRunes(8))).
		SetDescription("Test prompt for cost report tests").
		SetCreatorID(s.uid).
		SetProjectID(s.projectID).
		SaveX(context.Background())
	s.promptID = prompt.ID
}

func (s *costReportTestSuite) TestReports_Success() {
	// Create test prompt calls with different costs and dates
	now := time.Now()
	currentMonth := now.Format("2006-01")
	lastMonth := now.AddDate(0, -1, 0).Format("2006-01")

	// Create calls for current month
	s.createTestPromptCall(100.50, 1000, 800, now.AddDate(0, 0, -5), 0, false)
	s.createTestPromptCall(200.75, 1500, 1200, now.AddDate(0, 0, -3), 0, false)
	s.createTestPromptCall(50.25, 500, 400, now.AddDate(0, 0, -1), 0, true)

	// Create calls for last month
	lastMonthTime := now.AddDate(0, -1, 0)
	s.createTestPromptCall(300.00, 2000, 1600, lastMonthTime.AddDate(0, 0, -10), 0, false)
	s.createTestPromptCall(150.25, 1200, 1000, lastMonthTime.AddDate(0, 0, -5), 1, false)

	// Query cost reports for both months
	result, err := s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: lastMonth,
			To:   currentMonth,
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.reports)

	// Check count
	count := result.Count()
	assert.Equal(s.T(), int32(2), count)

	// Check reports
	reports := result.Reports()
	assert.Len(s.T(), reports, 2)

	// Verify current month report
	var currentMonthReport, lastMonthReport monthlyCostReportResponse
	for _, report := range reports {
		if report.Month() == currentMonth {
			currentMonthReport = report
		} else if report.Month() == lastMonth {
			lastMonthReport = report
		}
	}

	// Verify current month data
	assert.Equal(s.T(), currentMonth, currentMonthReport.Month())
	assert.Equal(s.T(), s.uid, currentMonthReport.UserId())
	assert.Equal(s.T(), 351.50, currentMonthReport.TotalCostCents()) // 100.50 + 200.75 + 50.25
	assert.Equal(s.T(), int32(3), currentMonthReport.TotalCalls())
	assert.Equal(s.T(), int32(3000), currentMonthReport.TotalTokens()) // 1000 + 1500 + 500
	assert.Equal(s.T(), int32(3), currentMonthReport.SuccessfulCalls()) // All have result = 0
	assert.Equal(s.T(), int32(1), currentMonthReport.CachedCalls()) // One cached call

	// Verify last month data
	assert.Equal(s.T(), lastMonth, lastMonthReport.Month())
	assert.Equal(s.T(), s.uid, lastMonthReport.UserId())
	assert.Equal(s.T(), 450.25, lastMonthReport.TotalCostCents()) // 300.00 + 150.25
	assert.Equal(s.T(), int32(2), lastMonthReport.TotalCalls())
	assert.Equal(s.T(), int32(3200), lastMonthReport.TotalTokens()) // 2000 + 1200
	assert.Equal(s.T(), int32(1), lastMonthReport.SuccessfulCalls()) // One success, one fail
	assert.Equal(s.T(), int32(0), lastMonthReport.CachedCalls()) // No cached calls

	// Verify breakdown data exists
	assert.NotEmpty(s.T(), currentMonthReport.CostsByProvider())
	assert.NotEmpty(s.T(), currentMonthReport.CostsByProject())
	assert.NotEmpty(s.T(), currentMonthReport.CostsByPrompt())
	assert.NotEmpty(s.T(), currentMonthReport.CostsByDay())
}

func (s *costReportTestSuite) TestReports_SingleMonth() {
	// Create test prompt calls for current month only
	now := time.Now()
	currentMonth := now.Format("2006-01")

	call1 := s.createTestPromptCall(125.75, 1000, 800, now.AddDate(0, 0, -5), 0, false)
	call2 := s.createTestPromptCall(275.25, 1500, 1200, now.AddDate(0, 0, -3), 0, false)

	// Query cost reports for current month only
	result, err := s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: currentMonth,
			To:   currentMonth,
		},
	})

	assert.Nil(s.T(), err)

	// Check count
	count := result.Count()
	assert.Equal(s.T(), int32(1), count)

	// Check reports
	reports := result.Reports()
	assert.Len(s.T(), reports, 1)

	report := reports[0]
	assert.Equal(s.T(), currentMonth, report.Month())
	assert.Equal(s.T(), 401.00, report.TotalCostCents()) // 125.75 + 275.25
	assert.Equal(s.T(), int32(2), report.TotalCalls())

	// Clean up
	service.EntClient.PromptCall.DeleteOneID(call1.ID).ExecX(context.Background())
	service.EntClient.PromptCall.DeleteOneID(call2.ID).ExecX(context.Background())
}

func (s *costReportTestSuite) TestReports_NoData() {
	// Query for a month with no data
	futureMonth := time.Now().AddDate(0, 2, 0).Format("2006-01")

	result, err := s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: futureMonth,
			To:   futureMonth,
		},
	})

	assert.Nil(s.T(), err)

	// Should return empty report for the month
	count := result.Count()
	assert.Equal(s.T(), int32(1), count)

	reports := result.Reports()
	assert.Len(s.T(), reports, 1)

	report := reports[0]
	assert.Equal(s.T(), futureMonth, report.Month())
	assert.Equal(s.T(), 0.0, report.TotalCostCents())
	assert.Equal(s.T(), int32(0), report.TotalCalls())
}

func (s *costReportTestSuite) TestReports_InvalidDateFormat() {
	// Test invalid from date
	_, err := s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: "2024-13", // Invalid month
			To:   "2024-01",
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)

	// Test invalid to date
	_, err = s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: "2024-01",
			To:   "invalid-date",
		},
	})

	assert.Error(s.T(), err)
	ge, ok = err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)
}

func (s *costReportTestSuite) TestReports_CostBreakdownItems() {
	// Create test prompt calls with different providers/projects
	now := time.Now()
	currentMonth := now.Format("2006-01")

	// Create a second project and prompt
	project2 := service.
		EntClient.
		Project.
		Create().
		SetName(fmt.Sprintf("Test Project 2 %s", utils.RandStringRunes(8))).
		SetCreatorID(s.uid).
		SaveX(context.Background())

	prompt2 := service.
		EntClient.
		Prompt.
		Create().
		SetName(fmt.Sprintf("Test Prompt 2 %s", utils.RandStringRunes(8))).
		SetDescription("Second test prompt").
		SetCreatorID(s.uid).
		SetProjectID(project2.ID).
		SaveX(context.Background())

	// Create calls for different projects and prompts
	call1 := s.createTestPromptCallWithProject(100.0, 1000, 800, now.AddDate(0, 0, -5), 0, false, s.projectID, s.promptID)
	call2 := s.createTestPromptCallWithProject(200.0, 1500, 1200, now.AddDate(0, 0, -3), 0, false, project2.ID, prompt2.ID)

	// Query cost reports
	result, err := s.q.Reports(s.ctx, reportsArgs{
		Filters: costReportFilters{
			From: currentMonth,
			To:   currentMonth,
		},
	})

	assert.Nil(s.T(), err)

	reports := result.Reports()
	assert.Len(s.T(), reports, 1)

	report := reports[0]

	// Check provider breakdown
	providerBreakdown := report.CostsByProvider()
	assert.Len(s.T(), providerBreakdown, 1) // All calls use same provider
	assert.Equal(s.T(), fmt.Sprintf("%d", s.provider.ID), providerBreakdown[0].ID())
	assert.Equal(s.T(), 300.0, providerBreakdown[0].CostCents())
	assert.Equal(s.T(), int32(2), providerBreakdown[0].Count())

	// Check project breakdown
	projectBreakdown := report.CostsByProject()
	assert.Len(s.T(), projectBreakdown, 2) // Two different projects

	// Check prompt breakdown
	promptBreakdown := report.CostsByPrompt()
	assert.Len(s.T(), promptBreakdown, 2) // Two different prompts

	// Check daily breakdown
	dailyBreakdown := report.CostsByDay()
	assert.Len(s.T(), dailyBreakdown, 2) // Two different days

	// Clean up
	service.EntClient.PromptCall.DeleteOneID(call1.ID).ExecX(context.Background())
	service.EntClient.PromptCall.DeleteOneID(call2.ID).ExecX(context.Background())
	service.EntClient.Prompt.DeleteOneID(prompt2.ID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(project2.ID).ExecX(context.Background())
}

func (s *costReportTestSuite) TestCostBreakdownItemResponse_AllFields() {
	item := service.CostBreakdownItem{
		ID:        "123",
		Name:      "Test Item",
		CostCents: 456.78,
		Count:     10,
	}

	response := costBreakdownItemResponse{item: item}

	assert.Equal(s.T(), "123", response.ID())
	assert.Equal(s.T(), "Test Item", response.Name())
	assert.Equal(s.T(), 456.78, response.CostCents())
	assert.Equal(s.T(), int32(10), response.Count())
}

func (s *costReportTestSuite) TestDailyCostResponse_AllFields() {
	cost := service.DailyCost{
		Date:      "2024-01-15",
		CostCents: 123.45,
		Count:     5,
	}

	response := dailyCostResponse{cost: cost}

	assert.Equal(s.T(), "2024-01-15", response.Date())
	assert.Equal(s.T(), 123.45, response.CostCents())
	assert.Equal(s.T(), int32(5), response.Count())
}

func (s *costReportTestSuite) TestMonthlyCostReportResponse_AllFields() {
	now := time.Now()
	report := service.MonthlyCostReport{
		Month:           "2024-01",
		UserID:          "test-user",
		TotalCostCents:  1234.56,
		TotalCalls:      100,
		TotalTokens:     50000,
		SuccessfulCalls: 95,
		CachedCalls:     10,
		CostsByProvider: []service.CostBreakdownItem{
			{ID: "1", Name: "Provider 1", CostCents: 600.0, Count: 60},
			{ID: "2", Name: "Provider 2", CostCents: 634.56, Count: 40},
		},
		CostsByProject: []service.CostBreakdownItem{
			{ID: "1", Name: "Project 1", CostCents: 1234.56, Count: 100},
		},
		CostsByPrompt: []service.CostBreakdownItem{
			{ID: "1", Name: "Prompt 1", CostCents: 800.0, Count: 70},
			{ID: "2", Name: "Prompt 2", CostCents: 434.56, Count: 30},
		},
		CostsByDay: []service.DailyCost{
			{Date: "2024-01-01", CostCents: 100.0, Count: 10},
			{Date: "2024-01-02", CostCents: 200.0, Count: 20},
		},
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
	}

	response := monthlyCostReportResponse{report: report}

	// Test basic fields
	assert.Equal(s.T(), "2024-01", response.Month())
	assert.Equal(s.T(), "test-user", response.UserId())
	assert.Equal(s.T(), 1234.56, response.TotalCostCents())
	assert.Equal(s.T(), int32(100), response.TotalCalls())
	assert.Equal(s.T(), int32(50000), response.TotalTokens())
	assert.Equal(s.T(), int32(95), response.SuccessfulCalls())
	assert.Equal(s.T(), int32(10), response.CachedCalls())

	// Test timestamps
	assert.Equal(s.T(), now.Format(time.RFC3339), response.CreatedAt())
	assert.Equal(s.T(), now.Add(time.Hour).Format(time.RFC3339), response.UpdatedAt())

	// Test breakdown arrays
	providers := response.CostsByProvider()
	assert.Len(s.T(), providers, 2)
	assert.Equal(s.T(), "Provider 1", providers[0].Name())

	projects := response.CostsByProject()
	assert.Len(s.T(), projects, 1)
	assert.Equal(s.T(), "Project 1", projects[0].Name())

	prompts := response.CostsByPrompt()
	assert.Len(s.T(), prompts, 2)
	assert.Equal(s.T(), "Prompt 1", prompts[0].Name())

	days := response.CostsByDay()
	assert.Len(s.T(), days, 2)
	assert.Equal(s.T(), "2024-01-01", days[0].Date())
}

// Helper method to create test prompt calls
func (s *costReportTestSuite) createTestPromptCall(costCents float64, totalTokens, responseTokens int, createdAt time.Time, result int, cached bool) *ent.PromptCall {
	return s.createTestPromptCallWithProject(costCents, totalTokens, responseTokens, createdAt, result, cached, s.projectID, s.promptID)
}

// Helper method to create test prompt calls with specific project and prompt
func (s *costReportTestSuite) createTestPromptCallWithProject(costCents float64, totalTokens, responseTokens int, createdAt time.Time, result int, cached bool, projectID, promptID int) *ent.PromptCall {
	call := service.EntClient.PromptCall.Create().
		SetPromptId(promptID).
		SetUserId(s.uid).
		SetResponseToken(responseTokens).
		SetTotalToken(totalTokens).
		SetDuration(1000).
		SetResult(result).
		SetCached(cached).
		SetCostCents(costCents).
		SetUa("PromptPal-Test/1.0").
		SetIP("127.0.0.1").
		SetProviderId(s.provider.ID).
		SetCreateTime(createdAt).
		SetUpdateTime(createdAt)

	// Set project relation
	created := call.SaveX(context.Background())

	// Add project relationship
	service.EntClient.Project.UpdateOneID(projectID).
		AddCalls(created).
		ExecX(context.Background())

	return created
}

func (s *costReportTestSuite) TearDownSuite() {
	// Clean up test data
	service.EntClient.Provider.DeleteOneID(s.provider.ID).ExecX(context.Background())
	service.EntClient.Prompt.DeleteOneID(s.promptID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(s.projectID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.uid).ExecX(context.Background())

	service.Close()
}

func TestCostReportTestSuite(t *testing.T) {
	suite.Run(t, new(costReportTestSuite))
}