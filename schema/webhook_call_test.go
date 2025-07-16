package schema

import (
	"context"
	"encoding/json"
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

type webhookCallTestSuite struct {
	suite.Suite
	uid       int
	projectID int
	webhookID int
	q         QueryResolver
	ctx       context.Context
}

func (s *webhookCallTestSuite) SetupSuite() {
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
	uniqueAddr := fmt.Sprintf("test-webhook-call-%s", utils.RandStringRunes(8))
	uniqueEmail := fmt.Sprintf("test-webhook-call-%s@annatarhe.com", utils.RandStringRunes(8))
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

	// Create test webhook
	webhook := service.
		EntClient.
		Webhook.
		Create().
		SetName(fmt.Sprintf("Test Webhook %s", utils.RandStringRunes(8))).
		SetDescription("Test webhook for webhook call tests").
		SetURL("https://example.com/webhook").
		SetEvent("onPromptFinished").
		SetEnabled(true).
		SetCreatorID(s.uid).
		SetProjectID(s.projectID).
		SaveX(context.Background())
	s.webhookID = webhook.ID
}

func (s *webhookCallTestSuite) TestWebhookCalls_Success() {
	// Configure mock expectations for RBAC permissions
	rbac := service.NewMockRBACService(s.T())
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	rbacService = rbac

	// Create test webhook calls
	call1 := s.createTestWebhookCall("trace-1", "https://example.com/webhook1", 200, false)
	call2 := s.createTestWebhookCall("trace-2", "https://example.com/webhook2", 500, false)
	call3 := s.createTestWebhookCall("trace-3", "https://example.com/webhook3", 0, true)

	// Query webhook calls
	result, err := s.q.WebhookCalls(s.ctx, webhookCallsArgs{
		Input: webhookCallsInput{
			WebhookID: int32(s.webhookID),
			Pagination: paginationInput{
				Limit:  10,
				Offset: 0,
			},
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.stat)

	// Check count
	count, err := result.Count(s.ctx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int32(3), count)

	// Check edges
	edges, err := result.Edges(s.ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), edges, 3)

	// Verify the calls are in descending order (most recent first)
	assert.Equal(s.T(), int32(call3.ID), edges[0].ID())
	assert.Equal(s.T(), int32(call2.ID), edges[1].ID())
	assert.Equal(s.T(), int32(call1.ID), edges[2].ID())

	// Clean up
	service.EntClient.WebhookCall.DeleteOneID(call1.ID).ExecX(context.Background())
	service.EntClient.WebhookCall.DeleteOneID(call2.ID).ExecX(context.Background())
	service.EntClient.WebhookCall.DeleteOneID(call3.ID).ExecX(context.Background())
}

func (s *webhookCallTestSuite) TestWebhookCalls_WithPagination() {
	// Configure mock expectations for RBAC permissions
	rbac := service.NewMockRBACService(s.T())
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	rbacService = rbac

	// Create multiple test webhook calls
	var calls []int
	for i := 0; i < 5; i++ {
		call := s.createTestWebhookCall(
			fmt.Sprintf("trace-%d", i),
			fmt.Sprintf("https://example.com/webhook%d", i),
			200,
			false,
		)
		calls = append(calls, call.ID)
		time.Sleep(1 * time.Millisecond) // Ensure different creation times
	}

	// Query with pagination: limit 2, offset 1
	result, err := s.q.WebhookCalls(s.ctx, webhookCallsArgs{
		Input: webhookCallsInput{
			WebhookID: int32(s.webhookID),
			Pagination: paginationInput{
				Limit:  2,
				Offset: 1,
			},
		},
	})

	assert.Nil(s.T(), err)

	// Check count (should be total count, not limited)
	count, err := result.Count(s.ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 5, count)

	// Check edges (should be 2 items with offset 1)
	edges, err := result.Edges(s.ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), edges, 2)

	// Clean up
	for _, callID := range calls {
		service.EntClient.WebhookCall.DeleteOneID(callID).ExecX(context.Background())
	}
}

func (s *webhookCallTestSuite) TestWebhookCalls_WebhookNotFound() {
	_, err := s.q.WebhookCalls(s.ctx, webhookCallsArgs{
		Input: webhookCallsInput{
			WebhookID: 99999,
			Pagination: paginationInput{
				Limit:  10,
				Offset: 0,
			},
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusNotFound, ge.code)
}

func (s *webhookCallTestSuite) TestWebhookCalls_InsufficientPermissions() {
	// Create context for different user (without permissions)
	rbac := service.NewMockRBACService(s.T())
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(false, nil)
	rbacService = rbac

	unauthorizedCtx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 99999, // Different user ID
	})

	_, err := s.q.WebhookCalls(unauthorizedCtx, webhookCallsArgs{
		Input: webhookCallsInput{
			WebhookID: int32(s.webhookID),
			Pagination: paginationInput{
				Limit:  10,
				Offset: 0,
			},
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusUnauthorized, ge.code)
}

func (s *webhookCallTestSuite) TestWebhookResponse_Calls() {
	// Create test webhook calls
	call1 := s.createTestWebhookCall("trace-1", "https://example.com/webhook1", 200, false)
	call2 := s.createTestWebhookCall("trace-2", "https://example.com/webhook2", 404, false)

	// Get webhook response
	webhook, err := service.EntClient.Webhook.Get(context.Background(), s.webhookID)
	assert.Nil(s.T(), err)
	webhookResp := webhookResponse{w: webhook}

	// Test calls method
	result, err := webhookResp.Calls(s.ctx, paginationInput{
		Limit:  10,
		Offset: 0,
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.stat)

	// Check count
	count, err := result.Count(s.ctx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int32(2), count)

	// Check edges
	edges, err := result.Edges(s.ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), edges, 2)

	// Clean up
	service.EntClient.WebhookCall.DeleteOneID(call1.ID).ExecX(context.Background())
	service.EntClient.WebhookCall.DeleteOneID(call2.ID).ExecX(context.Background())
}

func (s *webhookCallTestSuite) TestWebhookCallResponse_AllFields() {
	// Create a comprehensive webhook call
	requestHeaders := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "PromptPal-Webhook@test",
	}
	responseHeaders := map[string]string{
		"Content-Type": "application/json",
		"Server":       "nginx/1.18.0",
	}

	startTime := time.Now()
	endTime := startTime.Add(150 * time.Millisecond)

	call := service.EntClient.WebhookCall.Create().
		SetWebhookID(s.webhookID).
		SetTraceID("test-trace-comprehensive").
		SetURL("https://example.com/comprehensive").
		SetRequestHeaders(requestHeaders).
		SetRequestBody(`{"event":"onPromptFinished","data":"test"}`).
		SetStatusCode(201).
		SetResponseHeaders(responseHeaders).
		SetResponseBody(`{"success":true,"id":"12345"}`).
		SetStartTime(startTime).
		SetEndTime(endTime).
		SetIsTimeout(false).
		SetUserAgent("PromptPal-Webhook@test").
		SaveX(context.Background())

	callResp := webhookCallResponse{c: call}

	// Test all getter methods
	assert.Equal(s.T(), int32(call.ID), callResp.ID())
	assert.Equal(s.T(), int32(s.webhookID), callResp.WebhookID())
	assert.Equal(s.T(), "test-trace-comprehensive", callResp.TraceID())
	assert.Equal(s.T(), "https://example.com/comprehensive", callResp.URL())
	assert.Equal(s.T(), `{"event":"onPromptFinished","data":"test"}`, callResp.RequestBody())
	assert.Equal(s.T(), int32(201), *callResp.StatusCode())
	assert.Equal(s.T(), `{"success":true,"id":"12345"}`, *callResp.ResponseBody())
	assert.Equal(s.T(), startTime.Format(time.RFC3339), callResp.StartTime())
	assert.Equal(s.T(), endTime.Format(time.RFC3339), *callResp.EndTime())
	assert.False(s.T(), callResp.IsTimeout())
	assert.True(s.T(), callResp.IsSuccess())
	assert.Nil(s.T(), callResp.ErrorMessage())
	assert.Equal(s.T(), "PromptPal-Webhook@test", *callResp.UserAgent())

	// Test JSON headers
	requestHeadersJSON := callResp.RequestHeaders()
	assert.NotNil(s.T(), requestHeadersJSON)
	var parsedReqHeaders map[string]string
	err := json.Unmarshal(*requestHeadersJSON, &parsedReqHeaders)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), requestHeaders, parsedReqHeaders)

	responseHeadersJSON := callResp.ResponseHeaders()
	assert.NotNil(s.T(), responseHeadersJSON)
	var parsedRespHeaders map[string]string
	err = json.Unmarshal(*responseHeadersJSON, &parsedRespHeaders)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), responseHeaders, parsedRespHeaders)

	// Test webhook relationship
	webhookResp, err := callResp.Webhook(s.ctx)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int32(s.webhookID), webhookResp.ID())

	// Clean up
	service.EntClient.WebhookCall.DeleteOneID(call.ID).ExecX(context.Background())
}

func (s *webhookCallTestSuite) TestWebhookCallResponse_OptionalFields() {
	// Create a minimal webhook call (failed request)
	call := service.EntClient.WebhookCall.Create().
		SetWebhookID(s.webhookID).
		SetTraceID("test-trace-minimal").
		SetURL("https://example.com/minimal").
		SetRequestHeaders(map[string]string{"Content-Type": "application/json"}).
		SetRequestBody(`{"event":"onPromptFinished"}`).
		SetStartTime(time.Now()).
		SetIsTimeout(true).
		SetErrorMessage("Connection timeout").
		SaveX(context.Background())

	callResp := webhookCallResponse{c: call}

	// Test optional fields that should be nil/empty
	assert.Nil(s.T(), callResp.StatusCode())
	assert.Nil(s.T(), callResp.ResponseHeaders())
	assert.Nil(s.T(), callResp.ResponseBody())
	assert.Nil(s.T(), callResp.EndTime())
	assert.Nil(s.T(), callResp.UserAgent())

	// Test fields that should have values
	assert.True(s.T(), callResp.IsTimeout())
	assert.False(s.T(), callResp.IsSuccess())
	assert.Equal(s.T(), "Connection timeout", *callResp.ErrorMessage())

	// Clean up
	service.EntClient.WebhookCall.DeleteOneID(call.ID).ExecX(context.Background())
}

// Helper method to create test webhook calls
func (s *webhookCallTestSuite) createTestWebhookCall(traceID, url string, statusCode int, isTimeout bool) *ent.WebhookCall {
	call := service.EntClient.WebhookCall.Create().
		SetWebhookID(s.webhookID).
		SetTraceID(traceID).
		SetURL(url).
		SetRequestHeaders(map[string]string{"Content-Type": "application/json"}).
		SetRequestBody(`{"event":"onPromptFinished","test":true}`).
		SetStartTime(time.Now()).
		SetIsTimeout(isTimeout)

	if statusCode > 0 {
		call = call.SetStatusCode(statusCode)
	}

	// Determine success from status code (200-299)
	isSuccess := statusCode >= 200 && statusCode < 300

	if isSuccess {
		call = call.SetResponseBody(`{"success":true}`)
		call = call.SetResponseHeaders(map[string]string{"Content-Type": "application/json"})
	}
	if isTimeout {
		call = call.SetErrorMessage("Request timeout")
	} else if !isSuccess && statusCode > 0 {
		call = call.SetErrorMessage(fmt.Sprintf("HTTP %d error", statusCode))
	}

	return call.SaveX(context.Background())
}

func (s *webhookCallTestSuite) TearDownSuite() {
	// Clean up test data
	service.EntClient.Webhook.DeleteOneID(s.webhookID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(s.projectID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.uid).ExecX(context.Background())

	service.Close()
}

func TestWebhookCallTestSuite(t *testing.T) {
	suite.Run(t, new(webhookCallTestSuite))
}
