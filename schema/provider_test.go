package schema

import (
	"context"
	"net/http"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type providerTestSuite struct {
	suite.Suite
	uid       int
	projectID int
	promptID  int
	q         QueryResolver
	ctx       context.Context
}

func (s *providerTestSuite) SetupSuite() {
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
		SetAddr("test-addr-schema_provider_test005").
		SetName(utils.RandStringRunes(16)).
		SetLang("en").
		SetPhone(utils.RandStringRunes(16)).
		SetLevel(255).
		SetEmail("test-schema_provider_test005@annatarhe.com").
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
		SetName("Test Project").
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
		SetVariables([]dbSchema.PromptVariable{
			{Name: "variable1", Type: dbSchema.PromptVariableTypesString},
			{Name: "variable2", Type: dbSchema.PromptVariableTypesString},
		}).
		SaveX(context.Background())
	s.promptID = prompt.ID
}

func (s *providerTestSuite) TestCreateProvider_Success() {
	description := "Test provider for OpenAI"
	enabled := true
	temperature := 0.8
	topP := 0.95
	maxTokens := int32(2000)
	defaultModel := "gpt-4"
	organizationId := "org-123"

	result, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:           "Test OpenAI Provider",
			Description:    &description,
			Enabled:        &enabled,
			Source:         "openai",
			Endpoint:       "https://api.openai.com/v1",
			ApiKey:         "sk-test123",
			OrganizationId: &organizationId,
			DefaultModel:   &defaultModel,
			Temperature:    &temperature,
			TopP:           &topP,
			MaxTokens:      &maxTokens,
			Config:         `{"model":"gpt-4","stream":true}`,
			Headers:        `{"X-Custom":"test"}`,
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.p)
	assert.Equal(s.T(), "Test OpenAI Provider", result.Name())
	assert.Equal(s.T(), description, result.Description())
	assert.True(s.T(), result.Enabled())
	assert.Equal(s.T(), "openai", result.Source())
	assert.Equal(s.T(), "https://api.openai.com/v1", result.Endpoint())
	assert.Equal(s.T(), organizationId, *result.OrganizationId())
	assert.Equal(s.T(), defaultModel, result.DefaultModel())
	assert.Equal(s.T(), temperature, result.Temperature())
	assert.Equal(s.T(), topP, result.TopP())
	assert.Equal(s.T(), maxTokens, result.MaxTokens())
	assert.JSONEq(s.T(), `{"model":"gpt-4","stream":true}`, result.Config())
	assert.JSONEq(s.T(), `{"X-Custom":"test"}`, result.Headers())

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestCreateProvider_MinimalData() {
	result, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Minimal Provider",
			Source:   "gemini",
			Endpoint: "https://generativelanguage.googleapis.com",
			ApiKey:   "api-key-123",
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.p)
	assert.Equal(s.T(), "Minimal Provider", result.Name())
	assert.Equal(s.T(), "gemini", result.Source())
	assert.Equal(s.T(), "https://generativelanguage.googleapis.com", result.Endpoint())

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestCreateProvider_InvalidConfig() {
	_, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "sk-test123",
			Config:   `{"invalid": json}`, // Invalid JSON
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)
}

func (s *providerTestSuite) TestCreateProvider_InvalidHeaders() {
	_, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "sk-test123",
			Headers:  `{"invalid": json}`, // Invalid JSON
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)
}

func (s *providerTestSuite) TestUpdateProvider_Success() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Original Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "original-key",
		},
	})
	assert.Nil(s.T(), err)

	// Update the provider
	newName := "Updated Provider"
	newDescription := "Updated description"
	enabled := false
	newSource := "gemini"
	newEndpoint := "https://generativelanguage.googleapis.com"
	newApiKey := "new-api-key"
	newOrgId := "new-org-id"
	newModel := "gemini-pro"
	newTemp := 0.5
	newTopP := 0.8
	newMaxTokens := int32(1500)
	newConfig := `{"updated":true}`
	newHeaders := `{"Authorization":"Bearer new-token"}`

	result, err := s.q.UpdateProvider(s.ctx, updateProviderArgs{
		ID: provider.ID(),
		Data: updateProviderData{
			Name:           &newName,
			Description:    &newDescription,
			Enabled:        &enabled,
			Source:         &newSource,
			Endpoint:       &newEndpoint,
			ApiKey:         &newApiKey,
			OrganizationId: &newOrgId,
			DefaultModel:   &newModel,
			Temperature:    &newTemp,
			TopP:           &newTopP,
			MaxTokens:      &newMaxTokens,
			Config:         &newConfig,
			Headers:        &newHeaders,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), newName, result.Name())
	assert.Equal(s.T(), newDescription, result.Description())
	assert.False(s.T(), result.Enabled())
	assert.Equal(s.T(), newSource, result.Source())
	assert.Equal(s.T(), newEndpoint, result.Endpoint())
	assert.Equal(s.T(), newOrgId, *result.OrganizationId())
	assert.Equal(s.T(), newModel, result.DefaultModel())
	assert.Equal(s.T(), newTemp, result.Temperature())
	assert.Equal(s.T(), newTopP, result.TopP())
	assert.Equal(s.T(), newMaxTokens, result.MaxTokens())
	assert.JSONEq(s.T(), newConfig, result.Config())
	assert.JSONEq(s.T(), newHeaders, result.Headers())

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestUpdateProvider_PartialUpdate() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Original Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "original-key",
		},
	})
	assert.Nil(s.T(), err)

	// Update only the name
	newName := "Updated Name Only"
	result, err := s.q.UpdateProvider(s.ctx, updateProviderArgs{
		ID: provider.ID(),
		Data: updateProviderData{
			Name: &newName,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), newName, result.Name())
	assert.Equal(s.T(), "openai", result.Source()) // Should remain unchanged

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(result.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestUpdateProvider_NotFound() {
	newName := "Non-existent Provider"
	_, err := s.q.UpdateProvider(s.ctx, updateProviderArgs{
		ID: 99999,
		Data: updateProviderData{
			Name: &newName,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TestUpdateProvider_InvalidConfig() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	// Try to update with invalid config
	invalidConfig := `{"invalid": json}`
	_, err = s.q.UpdateProvider(s.ctx, updateProviderArgs{
		ID: provider.ID(),
		Data: updateProviderData{
			Config: &invalidConfig,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestUpdateProvider_InvalidHeaders() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	// Try to update with invalid headers
	invalidHeaders := `{"invalid": json}`
	_, err = s.q.UpdateProvider(s.ctx, updateProviderArgs{
		ID: provider.ID(),
		Data: updateProviderData{
			Headers: &invalidHeaders,
		},
	})

	assert.Error(s.T(), err)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusBadRequest, ge.code)

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestDeleteProvider_Success() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Provider to Delete",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	// Delete the provider
	result, err := s.q.DeleteProvider(s.ctx, deleteProviderArgs{
		ID: provider.ID(),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Verify provider is deleted
	_, err = service.EntClient.Provider.Get(context.Background(), int(provider.ID()))
	assert.Error(s.T(), err) // Should not be found
}

func (s *providerTestSuite) TestDeleteProvider_NotFound() {
	result, err := s.q.DeleteProvider(s.ctx, deleteProviderArgs{
		ID: 99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TestAssignProviderToProject_Success() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	// Assign provider to project
	result, err := s.q.AssignProviderToProject(s.ctx, assignProviderToProjectArgs{
		ProviderId: provider.ID(),
		ProjectId:  int32(s.projectID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Verify association exists
	providerResponse, err := s.q.ProjectProvider(s.ctx, projectProviderArgs{
		ProjectId: int32(s.projectID),
	})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), provider.ID(), providerResponse.ID())

	// Clean up
	s.q.RemoveProviderFromProject(s.ctx, removeProviderFromProjectArgs{ProjectId: int32(s.projectID)})
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestAssignProviderToProject_ProviderNotFound() {
	result, err := s.q.AssignProviderToProject(s.ctx, assignProviderToProjectArgs{
		ProviderId: 99999,
		ProjectId:  int32(s.projectID),
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TestAssignProviderToProject_ProjectNotFound() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	result, err := s.q.AssignProviderToProject(s.ctx, assignProviderToProjectArgs{
		ProviderId: provider.ID(),
		ProjectId:  99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestRemoveProviderFromProject_Success() {
	// First create a provider and assign to project
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	s.q.AssignProviderToProject(s.ctx, assignProviderToProjectArgs{
		ProviderId: provider.ID(),
		ProjectId:  int32(s.projectID),
	})

	// Remove provider from project
	result, err := s.q.RemoveProviderFromProject(s.ctx, removeProviderFromProjectArgs{
		ProjectId: int32(s.projectID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Verify association is removed
	providerResponse, err := s.q.ProjectProvider(s.ctx, projectProviderArgs{
		ProjectId: int32(s.projectID),
	})
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), providerResponse.p) // Should be empty

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestRemoveProviderFromProject_NoProviders() {
	result, err := s.q.RemoveProviderFromProject(s.ctx, removeProviderFromProjectArgs{
		ProjectId: int32(s.projectID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result) // Should return true even if no providers
}

func (s *providerTestSuite) TestRemoveProviderFromProject_ProjectNotFound() {
	result, err := s.q.RemoveProviderFromProject(s.ctx, removeProviderFromProjectArgs{
		ProjectId: 99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TestAssignProviderToPrompt_Success() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	// Assign provider to prompt
	result, err := s.q.AssignProviderToPrompt(s.ctx, assignProviderToPromptArgs{
		ProviderId: provider.ID(),
		PromptId:   int32(s.promptID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Clean up
	s.q.RemoveProviderFromPrompt(s.ctx, removeProviderFromPromptArgs{PromptId: int32(s.promptID)})
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestAssignProviderToPrompt_ProviderNotFound() {
	result, err := s.q.AssignProviderToPrompt(s.ctx, assignProviderToPromptArgs{
		ProviderId: 99999,
		PromptId:   int32(s.promptID),
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TestAssignProviderToPrompt_PromptNotFound() {
	// First create a provider
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	result, err := s.q.AssignProviderToPrompt(s.ctx, assignProviderToPromptArgs{
		ProviderId: provider.ID(),
		PromptId:   99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestRemoveProviderFromPrompt_Success() {
	// First create a provider and assign to prompt
	provider, err := s.q.CreateProvider(s.ctx, createProviderArgs{
		Data: createProviderData{
			Name:     "Test Provider",
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   "test-key",
		},
	})
	assert.Nil(s.T(), err)

	s.q.AssignProviderToPrompt(s.ctx, assignProviderToPromptArgs{
		ProviderId: provider.ID(),
		PromptId:   int32(s.promptID),
	})

	// Remove provider from prompt
	result, err := s.q.RemoveProviderFromPrompt(s.ctx, removeProviderFromPromptArgs{
		PromptId: int32(s.promptID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result)

	// Clean up
	service.EntClient.Provider.DeleteOneID(int(provider.ID())).ExecX(context.Background())
}

func (s *providerTestSuite) TestRemoveProviderFromPrompt_NoProviders() {
	result, err := s.q.RemoveProviderFromPrompt(s.ctx, removeProviderFromPromptArgs{
		PromptId: int32(s.promptID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), result) // Should return true even if no providers
}

func (s *providerTestSuite) TestRemoveProviderFromPrompt_PromptNotFound() {
	result, err := s.q.RemoveProviderFromPrompt(s.ctx, removeProviderFromPromptArgs{
		PromptId: 99999,
	})

	assert.Error(s.T(), err)
	assert.False(s.T(), result)
	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.Equal(s.T(), http.StatusInternalServerError, ge.code)
}

func (s *providerTestSuite) TearDownSuite() {
	// Clean up test data
	service.EntClient.Prompt.DeleteOneID(s.promptID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(s.projectID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.uid).ExecX(context.Background())

	service.Close()
}

func TestProviderTestSuite(t *testing.T) {
	suite.Run(t, new(providerTestSuite))
}
