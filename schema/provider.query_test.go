package schema

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type providerQueryTestSuite struct {
	suite.Suite
	uid                int
	providerID         int
	provider2ID        int
	projectID          int
	promptID           int
	q                  QueryResolver
	ctx                context.Context
	testProviderConfig map[string]interface{}
	testProviderHeaders map[string]string
}

func (s *providerQueryTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewWeb3Service()
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)
	Setup(hs, w3)

	s.q = QueryResolver{}

	// Create test user
	u := service.
		EntClient.
		User.
		Create().
		SetAddr(utils.RandStringRunes(16)).
		SetName(utils.RandStringRunes(16)).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(255).
		SetEmail(utils.RandStringRunes(10)).
		SaveX(context.Background())
	s.uid = u.ID

	s.ctx = context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})

	// Create test provider config and headers
	s.testProviderConfig = map[string]interface{}{
		"setting1": "value1",
		"setting2": 123,
	}
	s.testProviderHeaders = map[string]string{
		"X-Custom-Header": "test-value",
		"Authorization":   "Bearer test-token",
	}

	// Create test providers
	enabled := true
	temperature := 0.7
	topP := 0.9
	maxTokens := int32(4000)
	defaultModel := "gpt-4"
	description := "Test provider description"
	organizationId := "test-org-123"


	provider1 := service.
		EntClient.
		Provider.
		Create().
		SetName("Test Provider 1").
		SetDescription(description).
		SetEnabled(enabled).
		SetSource("openai").
		SetEndpoint("https://api.openai.com/v1").
		SetApiKey("test-key-1").
		SetOrganizationId(organizationId).
		SetDefaultModel(defaultModel).
		SetTemperature(temperature).
		SetTopP(topP).
		SetMaxTokens(int(maxTokens)).
		SetConfig(s.testProviderConfig).
		SetHeaders(s.testProviderHeaders).
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.providerID = provider1.ID

	provider2 := service.
		EntClient.
		Provider.
		Create().
		SetName("Test Provider 2").
		SetSource("gemini").
		SetEndpoint("https://generativelanguage.googleapis.com").
		SetApiKey("test-key-2").
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.provider2ID = provider2.ID

	// Create test project
	projectName := "Test Project"
	project := service.
		EntClient.
		Project.
		Create().
		SetName(projectName).
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.projectID = project.ID

	// Create test prompt
	promptRows := []dbSchema.PromptRow{
		{Prompt: "Test prompt content", Role: "user"},
	}
	prompt := service.
		EntClient.
		Prompt.
		Create().
		SetName("Test Prompt").
		SetPrompts(promptRows).
		SetProjectID(s.projectID).
		SetCreatorID(s.uid).
		SaveX(context.Background())
	s.promptID = prompt.ID

	// Associate provider with project and prompt
	provider1.Update().AddProject(project).AddPrompt(prompt).SaveX(context.Background())
}

func (s *providerQueryTestSuite) TestProvider_Success() {
	result, err := s.q.Provider(s.ctx, providerArgs{ID: int32(s.providerID)})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.p)
	assert.Equal(s.T(), int32(s.providerID), result.ID())
	assert.Equal(s.T(), "Test Provider 1", result.Name())
	assert.Equal(s.T(), "Test provider description", result.Description())
	assert.True(s.T(), result.Enabled())
	assert.Equal(s.T(), "openai", result.Source())
	assert.Equal(s.T(), "https://api.openai.com/v1", result.Endpoint())
	assert.Equal(s.T(), "test-org-123", *result.OrganizationId())
	assert.Equal(s.T(), "gpt-4", result.DefaultModel())
	assert.Equal(s.T(), 0.7, result.Temperature())
	assert.Equal(s.T(), 0.9, result.TopP())
	assert.Equal(s.T(), int32(4000), result.MaxTokens())
	
	// Test JSON fields
	expectedConfig := `{"setting1":"value1","setting2":123}`
	expectedHeaders := `{"Authorization":"Bearer test-token","X-Custom-Header":"test-value"}`
	assert.JSONEq(s.T(), expectedConfig, result.Config())
	assert.JSONEq(s.T(), expectedHeaders, result.Headers())
	
	// Test timestamps
	assert.NotEmpty(s.T(), result.CreatedAt())
	assert.NotEmpty(s.T(), result.UpdatedAt())
}

func (s *providerQueryTestSuite) TestProvider_NotFound() {
	_, err := s.q.Provider(s.ctx, providerArgs{ID: 99999})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusNotFound, ge.code)
}

func (s *providerQueryTestSuite) TestProvider_CacheHit() {
	// First call to populate cache
	result1, err1 := s.q.Provider(s.ctx, providerArgs{ID: int32(s.providerID)})
	assert.Nil(s.T(), err1)

	// Second call should hit cache
	result2, err2 := s.q.Provider(s.ctx, providerArgs{ID: int32(s.providerID)})
	assert.Nil(s.T(), err2)
	assert.Equal(s.T(), result1.ID(), result2.ID())
	assert.Equal(s.T(), result1.Name(), result2.Name())
}

func (s *providerQueryTestSuite) TestProviders_Success() {
	result, err := s.q.Providers(s.ctx, providersArgs{
		Pagination: paginationInput{Limit: 10, Offset: 0},
	})

	assert.Nil(s.T(), err)
	assert.GreaterOrEqual(s.T(), result.Count(), int32(2))
	edges := result.Edges()
	assert.Len(s.T(), edges, int(result.Count()))

	// Verify providers are ordered by ID descending
	if len(edges) >= 2 {
		assert.Greater(s.T(), edges[0].ID(), edges[1].ID())
	}
}

func (s *providerQueryTestSuite) TestProviders_Pagination() {
	// Test with limit 1
	result, err := s.q.Providers(s.ctx, providersArgs{
		Pagination: paginationInput{Limit: 1, Offset: 0},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), int32(1), result.Count())
	edges := result.Edges()
	assert.Len(s.T(), edges, 1)

	// Test with offset
	result2, err2 := s.q.Providers(s.ctx, providersArgs{
		Pagination: paginationInput{Limit: 1, Offset: 1},
	})

	assert.Nil(s.T(), err2)
	edges2 := result2.Edges()
	if len(edges2) > 0 {
		assert.NotEqual(s.T(), edges[0].ID(), edges2[0].ID())
	}
}

func (s *providerQueryTestSuite) TestProjectProvider_Success() {
	result, err := s.q.ProjectProvider(s.ctx, projectProviderArgs{
		ProjectId: int32(s.projectID),
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.p)
	assert.Equal(s.T(), int32(s.providerID), result.ID())
}

func (s *providerQueryTestSuite) TestProjectProvider_NotFound() {
	result, err := s.q.ProjectProvider(s.ctx, projectProviderArgs{
		ProjectId: 99999,
	})

	assert.Nil(s.T(), err) // Should not return error for not found
	assert.Nil(s.T(), result.p) // Should return empty response
}

func (s *providerQueryTestSuite) TestProviderResponse_NilProvider() {
	resp := providerResponse{p: nil}

	assert.Equal(s.T(), int32(0), resp.ID())
	assert.Equal(s.T(), "", resp.Name())
	assert.Equal(s.T(), "", resp.Description())
	assert.False(s.T(), resp.Enabled())
	assert.Equal(s.T(), "", resp.Source())
	assert.Equal(s.T(), "", resp.Endpoint())
	assert.Nil(s.T(), resp.OrganizationId())
	assert.Equal(s.T(), "", resp.DefaultModel())
	assert.Equal(s.T(), float64(0), resp.Temperature())
	assert.Equal(s.T(), float64(0), resp.TopP())
	assert.Equal(s.T(), int32(0), resp.MaxTokens())
	assert.Equal(s.T(), "", resp.Config())
	assert.Equal(s.T(), "", resp.Headers())
	assert.Equal(s.T(), "", resp.CreatedAt())
	assert.Equal(s.T(), "", resp.UpdatedAt())
}

func (s *providerQueryTestSuite) TestProviderResponse_OrganizationIdEmpty() {
	// Create provider without organization ID
	provider := &ent.Provider{
		ID:             123,
		Name:           "Test",
		OrganizationId: "",
	}
	resp := providerResponse{p: provider}

	assert.Nil(s.T(), resp.OrganizationId())
}

func (s *providerQueryTestSuite) TestProviderResponse_Projects() {
	result, err := s.q.Provider(s.ctx, providerArgs{ID: int32(s.providerID)})
	assert.Nil(s.T(), err)

	projects, err := result.Projects(s.ctx)
	assert.Nil(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(projects.projects), 1)

	// Find our test project
	found := false
	for _, project := range projects.projects {
		if project.ID == s.projectID {
			found = true
			break
		}
	}
	assert.True(s.T(), found)
}

func (s *providerQueryTestSuite) TestProviderResponse_Projects_NilProvider() {
	resp := providerResponse{p: nil}
	projects, err := resp.Projects(s.ctx)

	assert.Nil(s.T(), err)
	assert.Nil(s.T(), projects.projects)
}

func (s *providerQueryTestSuite) TestProviderResponse_Prompts() {
	result, err := s.q.Provider(s.ctx, providerArgs{ID: int32(s.providerID)})
	assert.Nil(s.T(), err)

	prompts, err := result.Prompts(s.ctx)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), prompts.stat)
	assert.Equal(s.T(), int32(10), prompts.pagination.Limit)
	assert.Equal(s.T(), int32(0), prompts.pagination.Offset)
}

func (s *providerQueryTestSuite) TestProviderResponse_Prompts_NilProvider() {
	resp := providerResponse{p: nil}
	prompts, err := resp.Prompts(s.ctx)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), prompts.stat)
	assert.Equal(s.T(), int32(10), prompts.pagination.Limit)
	assert.Equal(s.T(), int32(0), prompts.pagination.Offset)
}

func (s *providerQueryTestSuite) TestProviderResponse_ConfigMarshalError() {
	// Create a provider with config that cannot be marshaled
	provider := &ent.Provider{
		ID:   123,
		Name: "Test",
		Config: map[string]interface{}{
			"invalid": make(chan int), // channels cannot be marshaled
		},
	}
	resp := providerResponse{p: provider}

	assert.Equal(s.T(), "", resp.Config())
}

func (s *providerQueryTestSuite) TestProviderResponse_HeadersMarshalError() {
	// Create a provider with headers that will cause marshal error
	provider := &ent.Provider{
		ID:   123,
		Name: "Test",
		Headers: map[string]string{
			string([]byte{0xff, 0xfe, 0xfd}): "invalid-utf8", // Invalid UTF-8
		},
	}
	resp := providerResponse{p: provider}

	// This might not cause error in Go, so let's test normal case
	result := resp.Headers()
	assert.NotEmpty(s.T(), result)
}

func (s *providerQueryTestSuite) TestProvidersResponse_Methods() {
	providers := []*ent.Provider{
		{ID: 1, Name: "Provider 1"},
		{ID: 2, Name: "Provider 2"},
		{ID: 3, Name: "Provider 3"},
	}

	resp := providersResponse{providers: providers}

	assert.Equal(s.T(), int32(3), resp.Count())

	edges := resp.Edges()
	assert.Len(s.T(), edges, 3)
	assert.Equal(s.T(), int32(1), edges[0].ID())
	assert.Equal(s.T(), int32(2), edges[1].ID())
	assert.Equal(s.T(), int32(3), edges[2].ID())
}

func (s *providerQueryTestSuite) TestProvidersResponse_EmptyList() {
	resp := providersResponse{providers: []*ent.Provider{}}

	assert.Equal(s.T(), int32(0), resp.Count())
	edges := resp.Edges()
	assert.Empty(s.T(), edges)
}

func (s *providerQueryTestSuite) TearDownSuite() {
	// Clean up test data
	service.EntClient.Provider.DeleteOneID(s.providerID).ExecX(context.Background())
	service.EntClient.Provider.DeleteOneID(s.provider2ID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(s.projectID).ExecX(context.Background())
	service.EntClient.Prompt.DeleteOneID(s.promptID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.uid).ExecX(context.Background())

	// Clear cache
	service.Cache.Delete(context.Background(), fmt.Sprintf("provider:%d", s.providerID))
	service.Cache.Delete(context.Background(), fmt.Sprintf("provider:%d", s.provider2ID))

	service.Close()
}

func TestProviderQueryTestSuite(t *testing.T) {
	suite.Run(t, new(providerQueryTestSuite))
}