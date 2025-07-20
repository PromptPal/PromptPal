package schema

import (
	"context"
	"net/http"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type webhookTestSuite struct {
	suite.Suite
	uid       int
	projectID int
	q         QueryResolver
	ctx       context.Context
}

func (s *webhookTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewWeb3Service()
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)

	rbac := service.NewMockRBACService(s.T())
	// Configure mock expectations for RBAC permissions
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	Setup(hs, w3, rbac)

	s.q = QueryResolver{}

	// Create test user with unique identifier
	testUserName := "test-user-webhook-" + utils.RandStringRunes(8)
	testUserAddr := "test-addr-webhook-" + utils.RandStringRunes(8)
	testUserEmail := testUserAddr + "@test-webhook.com"

	u := service.
		EntClient.
		User.
		Create().
		SetAddr(testUserAddr).
		SetName(testUserName).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(255).
		SetEmail(testUserEmail).
		SaveX(context.Background())
	s.uid = u.ID

	s.ctx = context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})

	// Create test project with unique name
	testProjectName := "Test Webhook Project " + utils.RandStringRunes(8)
	project := service.
		EntClient.
		Project.
		Create().
		SetName(testProjectName).
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.projectID = project.ID
}

func (s *webhookTestSuite) TestCreateWebhook_Success() {
	description := "Test webhook description"
	enabled := true

	result, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:        "Test Webhook",
			Description: &description,
			URL:         "https://api.example.com/webhook",
			Event:       EventOnPromptFinished,
			Enabled:     &enabled,
			ProjectID:   int32(s.projectID),
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.w)
	assert.Equal(s.T(), "Test Webhook", result.Name())
	assert.Equal(s.T(), description, result.Description())
	assert.Equal(s.T(), "https://api.example.com/webhook", result.URL())
	assert.Equal(s.T(), EventOnPromptFinished, result.Event())
	assert.True(s.T(), result.Enabled())

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestCreateWebhook_MinimalData() {
	result, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Minimal Webhook",
			URL:       "https://minimal.example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.w)
	assert.Equal(s.T(), "Minimal Webhook", result.Name())
	assert.Equal(s.T(), "https://minimal.example.com/webhook", result.URL())
	assert.Equal(s.T(), EventOnPromptFinished, result.Event())
	assert.True(s.T(), result.Enabled()) // Default should be true

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestCreateWebhook_InvalidEvent() {
	_, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Invalid Event Webhook",
			URL:       "https://example.com/webhook",
			Event:     "invalidEvent",
			ProjectID: int32(s.projectID),
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)
	assert.Contains(s.T(), err.Error(), "only onPromptFinished event is supported")
}

func (s *webhookTestSuite) TestCreateWebhook_InvalidURL() {
	testCases := []struct {
		name        string
		url         string
		expectedErr string
	}{
		{
			name:        "empty URL",
			url:         "",
			expectedErr: "URL cannot be empty",
		},
		{
			name:        "invalid scheme",
			url:         "ftp://example.com/webhook",
			expectedErr: "only HTTP and HTTPS schemes are allowed",
		},
		{
			name:        "localhost URL",
			url:         "http://localhost:8080/webhook",
			expectedErr: "localhost URLs are not allowed",
		},
		{
			name:        "private IP",
			url:         "http://192.168.1.1/webhook",
			expectedErr: "private, loopback, and link-local IPs are not allowed",
		},
		{
			name:        "loopback IP",
			url:         "http://127.0.0.1/webhook",
			expectedErr: "private, loopback, and link-local IPs are not allowed",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
				Data: createWebhookData{
					Name:      "Invalid URL Webhook",
					URL:       tc.url,
					Event:     EventOnPromptFinished,
					ProjectID: int32(s.projectID),
				},
			})

			assert.Error(t, err)
			ge, ok := err.(GraphQLHttpError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusBadRequest, ge.code)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func (s *webhookTestSuite) TestUpdateWebhook_Success() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Original Webhook",
			URL:       "https://original.example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Update the webhook
	newName := "Updated Webhook"
	newDescription := "Updated description"
	enabled := false
	newURL := "https://updated.example.com/webhook"

	result, err := s.q.UpdateWebhook(s.ctx, updateWebhookArgs{
		ID: webhook.ID(),
		Data: updateWebhookData{
			Name:        &newName,
			Description: &newDescription,
			Enabled:     &enabled,
			URL:         &newURL,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), newName, result.Name())
	assert.Equal(s.T(), newDescription, result.Description())
	assert.False(s.T(), result.Enabled())
	assert.Equal(s.T(), newURL, result.URL())

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestUpdateWebhook_PartialUpdate() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Original Webhook",
			URL:       "https://original.example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Update only the name
	newName := "Updated Name Only"
	result, err := s.q.UpdateWebhook(s.ctx, updateWebhookArgs{
		ID: webhook.ID(),
		Data: updateWebhookData{
			Name: &newName,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), newName, result.Name())
	assert.Equal(s.T(), "https://original.example.com/webhook", result.URL()) // Should remain unchanged

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestUpdateWebhook_NotFound() {
	newName := "Non-existent Webhook"
	_, err := s.q.UpdateWebhook(s.ctx, updateWebhookArgs{
		ID: 99999,
		Data: updateWebhookData{
			Name: &newName,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusNotFound, ge.code)
}

func (s *webhookTestSuite) TestUpdateWebhook_InvalidEvent() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Test Webhook",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Try to update with invalid event
	invalidEvent := "invalidEvent"
	_, err = s.q.UpdateWebhook(s.ctx, updateWebhookArgs{
		ID: webhook.ID(),
		Data: updateWebhookData{
			Event: &invalidEvent,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(webhook.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestUpdateWebhook_InvalidURL() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Test Webhook",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Try to update with invalid URL
	invalidURL := "http://localhost/webhook"
	_, err = s.q.UpdateWebhook(s.ctx, updateWebhookArgs{
		ID: webhook.ID(),
		Data: updateWebhookData{
			URL: &invalidURL,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(webhook.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestDeleteWebhook_Success() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Webhook to Delete",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Delete the webhook
	result, err := s.q.DeleteWebhook(s.ctx, deleteWebhookArgs{
		ID: webhook.ID(),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Verify webhook is deleted
	_, err = service.EntClient.Webhook.Get(context.Background(), int(webhook.ID()))
	assert.Error(s.T(), err) // Should not be found
}

func (s *webhookTestSuite) TestDeleteWebhook_NotFound() {
	result, err := s.q.DeleteWebhook(s.ctx, deleteWebhookArgs{
		ID: 99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusNotFound, ge.code)
}

func (s *webhookTestSuite) TestWebhook_Success() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Single Test Webhook",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Get single webhook by ID
	result, err := s.q.Webhook(s.ctx, webhookArgs{
		ID: webhook.ID(),
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.w)
	assert.Equal(s.T(), webhook.ID(), result.ID())
	assert.Equal(s.T(), "Single Test Webhook", result.Name())
	assert.Equal(s.T(), "https://example.com/webhook", result.URL())
	assert.Equal(s.T(), EventOnPromptFinished, result.Event())

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(webhook.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestWebhook_NotFound() {
	// Try to get non-existent webhook
	_, err := s.q.Webhook(s.ctx, webhookArgs{
		ID: 99999,
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusNotFound, ge.code)
}

func (s *webhookTestSuite) TestWebhook_UnauthorizedAccess() {
	// Create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "Unauthorized Test Webhook",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// Create another user without permissions
	testUserName := "unauthorized-user-webhook-single-" + utils.RandStringRunes(8)
	testUserAddr := "unauthorized-addr-webhook-single-" + utils.RandStringRunes(8)
	testUserEmail := testUserAddr + "@test-webhook.com"

	unauthorizedUser := service.
		EntClient.
		User.
		Create().
		SetAddr(testUserAddr).
		SetName(testUserName).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(0). // No admin level
		SetEmail(testUserEmail).
		SaveX(context.Background())

	unauthorizedCtx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: unauthorizedUser.ID,
	})

	// Mock RBAC to return false for unauthorized user viewing, but allow other operations
	rbac := service.NewMockRBACService(s.T())
	rbac.On("HasPermission", mock.Anything, unauthorizedUser.ID, mock.Anything, service.PermProjectView).Return(false, nil)
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, service.PermProjectEdit).Return(true, nil)
	rbacService = rbac

	// Try to get webhook without permissions
	_, err = s.q.Webhook(unauthorizedCtx, webhookArgs{
		ID: webhook.ID(),
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusUnauthorized, ge.code)

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(webhook.ID())).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(unauthorizedUser.ID).ExecX(context.Background())
}

func (s *webhookTestSuite) TestWebhooks_Success() {
	// First create a webhook
	webhook, err := s.q.CreateWebhook(s.ctx, createWebhookArgs{
		Data: createWebhookData{
			Name:      "List Test Webhook",
			URL:       "https://example.com/webhook",
			Event:     EventOnPromptFinished,
			ProjectID: int32(s.projectID),
		},
	})
	assert.Nil(s.T(), err)

	// List webhooks
	result, err := s.q.Webhooks(s.ctx, webhooksArgs{
		ProjectID: int32(s.projectID),
		Pagination: paginationInput{
			Limit:  10,
			Offset: 0,
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.stat)

	count, err := result.Count(s.ctx)
	assert.Nil(s.T(), err)
	assert.Greater(s.T(), count, int32(0))

	edges, err := result.Edges(s.ctx)
	assert.Nil(s.T(), err)
	assert.Greater(s.T(), len(edges), 0)

	found := false
	for _, edge := range edges {
		if edge.Name() == "List Test Webhook" {
			found = true
			break
		}
	}
	assert.True(s.T(), found)

	// Clean up
	service.EntClient.Webhook.DeleteOneID(int(webhook.ID())).ExecX(context.Background())
}

func (s *webhookTestSuite) TestWebhooks_UnauthorizedAccess() {
	// Create another user without permissions
	testUserName := "unauthorized-user-webhook-" + utils.RandStringRunes(8)
	testUserAddr := "unauthorized-addr-webhook-" + utils.RandStringRunes(8)
	testUserEmail := testUserAddr + "@test-webhook.com"

	unauthorizedUser := service.
		EntClient.
		User.
		Create().
		SetAddr(testUserAddr).
		SetName(testUserName).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(0). // No admin level
		SetEmail(testUserEmail).
		SaveX(context.Background())

	unauthorizedCtx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: unauthorizedUser.ID,
	})

	// Mock RBAC to return false for unauthorized user viewing, but allow other operations
	rbac := service.NewMockRBACService(s.T())
	rbac.On("HasPermission", mock.Anything, unauthorizedUser.ID, mock.Anything, service.PermProjectView).Return(false, nil)
	rbac.On("HasPermission", mock.Anything, mock.Anything, mock.Anything, service.PermProjectEdit).Return(true, nil)
	rbacService = rbac

	// Try to list webhooks without permissions
	_, err := s.q.Webhooks(unauthorizedCtx, webhooksArgs{
		ProjectID: int32(s.projectID),
		Pagination: paginationInput{
			Limit:  10,
			Offset: 0,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusUnauthorized, ge.code)

	// Clean up
	service.EntClient.User.DeleteOneID(unauthorizedUser.ID).ExecX(context.Background())
}

func (s *webhookTestSuite) TestWebhookResponse_Methods() {
	// Create test webhook directly in database
	testWebhookName := "Response Test Webhook " + utils.RandStringRunes(8)
	webhook := service.
		EntClient.
		Webhook.
		Create().
		SetName(testWebhookName).
		SetDescription("Test Description").
		SetURL("https://example.com/webhook").
		SetEvent(EventOnPromptFinished).
		SetEnabled(true).
		SetCreatorID(s.uid).
		SetProjectID(s.projectID).
		SaveX(context.Background())

	resp := webhookResponse{w: webhook}

	assert.Equal(s.T(), int32(webhook.ID), resp.ID())
	assert.Equal(s.T(), testWebhookName, resp.Name())
	assert.Equal(s.T(), "Test Description", resp.Description())
	assert.Equal(s.T(), "https://example.com/webhook", resp.URL())
	assert.Equal(s.T(), EventOnPromptFinished, resp.Event())
	assert.True(s.T(), resp.Enabled())
	assert.NotEmpty(s.T(), resp.CreatedAt())
	assert.NotEmpty(s.T(), resp.UpdatedAt())

	// Test creator relation
	creator, err := resp.Creator(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.uid, creator.u.ID)

	// Test project relation
	proj, err := resp.Project(context.Background())
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.projectID, proj.p.ID)

	// Clean up
	service.EntClient.Webhook.DeleteOneID(webhook.ID).ExecX(context.Background())
}

func (s *webhookTestSuite) TearDownSuite() {
	// Clean up test data
	service.EntClient.Project.DeleteOneID(s.projectID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.uid).ExecX(context.Background())

	service.Close()
}

func TestWebhookTestSuite(t *testing.T) {
	suite.Run(t, new(webhookTestSuite))
}
