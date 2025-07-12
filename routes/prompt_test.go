package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/provider"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type promptTestSuite struct {
	suite.Suite
	router   *gin.Engine
	w3       *service.MockWeb3Service
	iai      *service.MockIsomorphicAIService
	hashid   *service.MockHashIDService
	user     *ent.User
	project  *ent.Project
	provider *ent.Provider
}

func (s *promptTestSuite) SetupTest() {
	config.SetupConfig(true)
	s.w3 = service.NewMockWeb3Service(s.T())
	s.iai = service.NewMockIsomorphicAIService(s.T())
	s.hashid = service.NewMockHashIDService(s.T())

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)

	// Initialize minimal cache for testing
	cache := cache.New(&cache.Options{
		LocalCache: cache.NewTinyLFU(100, time.Minute),
	})
	service.Cache = cache

	// Initialize services for route handlers
	web3Service = s.w3
	isomorphicAIService = s.iai
	hashidService = s.hashid

	// Create minimal gin router for testing
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	// Create test data
	s.createTestData()
}

func (s *promptTestSuite) createTestData() {
	// Create test user
	user, err := service.EntClient.User.
		Create().
		SetUsername("testuser").
		SetEmail("test@example.com").
		SetAddr("0x1234567890123456789012345678901234567890").
		SetName("Test User").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())
	assert.Nil(s.T(), err)
	s.user = user

	// Create test provider
	provider, err := service.EntClient.Provider.
		Create().
		SetName("Test Provider").
		SetSource("openai").
		SetApiKey("test-key").
		SetDefaultModel("gpt-3.5-turbo").
		SetTemperature(0.7).
		SetTopP(1.0).
		SetMaxTokens(2048).
		SetCreatorID(user.ID).
		Save(context.Background())
	assert.Nil(s.T(), err)
	s.provider = provider

	// Create test project
	project, err := service.EntClient.Project.
		Create().
		SetName("Test Project").
		SetCreatorID(user.ID).
		SetProviderId(provider.ID).
		SetOpenAIBaseURL("https://api.openai.com/v1").
		SetOpenAIToken("test-token").
		SetOpenAIModel("gpt-3.5-turbo").
		SetOpenAITemperature(0.7).
		SetOpenAITopP(1.0).
		SetOpenAIMaxTokens(2048).
		Save(context.Background())
	assert.Nil(s.T(), err)
	s.project = project
}

func (s *promptTestSuite) getAuthHeaders() map[string]string {
	// Mock JWT auth - in real scenarios would use actual JWT
	// For testing, we'll rely on test mode setup
	return map[string]string{
		"Authorization": "Bearer test-token",
	}
}

func (s *promptTestSuite) TestTestPrompt() {
	// Create test payload
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello {{name}}",
			},
		},
		Variables: map[string]string{
			"name": "John",
		},
	}

	// Mock expected response
	mockResponse := openai.ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-3.5-turbo",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello John",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	// Set up mock expectations
	s.iai.On("Chat",
		mock.AnythingOfType("context.backgroundCtx"), // context
		mock.MatchedBy(func(p *ent.Provider) bool {
			return p.ID == s.provider.ID && p.Name == "Test Provider"
		}), // provider
		mock.MatchedBy(func(prompt ent.Prompt) bool {
			return len(prompt.Prompts) == 1 && prompt.Prompts[0].Role == "user"
		}), // prompt
		payload.Variables, // variables
		"",                // userId (empty for test prompt)
	).Return(mockResponse, nil)

	// Prepare request
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add auth headers
	for k, v := range s.getAuthHeaders() {
		req.Header.Set(k, v)
	}

	// Create gin context with authenticated user
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	// Call the handler
	testPrompt(c)

	// Assert response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response openai.ChatCompletionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "chatcmpl-123", response.ID)
	assert.Equal(s.T(), "Hello John", response.Choices[0].Message.Content)
	assert.Equal(s.T(), 15, response.Usage.TotalTokens)

	// Verify all expectations were met
	s.iai.AssertExpectations(s.T())
}

func (s *promptTestSuite) TestTestPromptUnauthorized() {
	// Test with uid = 0 (unauthorized)
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello",
			},
		},
		Variables: map[string]string{},
	}

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", 0) // Unauthorized user

	testPrompt(c)

	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)

	var response errorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid uid", response.ErrorMessage)
}

func (s *promptTestSuite) TestTestPromptInvalidJSON() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)

	var response errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), response.ErrorMessage, "invalid character")
}

func (s *promptTestSuite) TestTestPromptMissingRequiredFields() {
	// Test with missing projectId (required field)
	payload := testPromptPayload{
		// ProjectID missing
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello",
			},
		},
		Variables: map[string]string{},
	}

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *promptTestSuite) TestTestPromptProviderNotFound() {
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: 99999, // Non-existent provider ID
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello",
			},
		},
		Variables: map[string]string{},
	}

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response errorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Contains(s.T(), response.ErrorMessage, "not found")
}

func (s *promptTestSuite) TestTestPromptChatError() {
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello {{name}}",
			},
		},
		Variables: map[string]string{
			"name": "John",
		},
	}

	// Mock chat service to return error
	s.iai.On("Chat",
		mock.AnythingOfType("context.backgroundCtx"), // context
		mock.MatchedBy(func(p *ent.Provider) bool {
			return p.ID == s.provider.ID && p.Name == "Test Provider"
		}), // provider
		mock.MatchedBy(func(prompt ent.Prompt) bool {
			return len(prompt.Prompts) == 1 && prompt.Prompts[0].Role == "user"
		}), // prompt
		payload.Variables, // variables
		"",                // userId
	).Return(openai.ChatCompletionResponse{}, assert.AnError)

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)

	var response errorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), assert.AnError.Error(), response.ErrorMessage)

	// Verify all expectations were met
	s.iai.AssertExpectations(s.T())
}

func (s *promptTestSuite) TestTestPromptWithVariableSubstitution() {
	// Test with multiple variables in prompt
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "user",
				Prompt: "Hello {{name}}, you are {{age}} years old and live in {{city}}",
			},
		},
		Variables: map[string]string{
			"name": "Alice",
			"age":  "25",
			"city": "New York",
		},
	}

	mockResponse := openai.ChatCompletionResponse{
		ID:      "chatcmpl-456",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-3.5-turbo",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "Hello Alice! Nice to meet you.",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     20,
			CompletionTokens: 8,
			TotalTokens:      28,
		},
	}

	s.iai.On("Chat",
		mock.AnythingOfType("context.backgroundCtx"),
		mock.MatchedBy(func(p *ent.Provider) bool {
			return p.ID == s.provider.ID && p.Name == "Test Provider"
		}),
		mock.MatchedBy(func(prompt ent.Prompt) bool {
			return len(prompt.Prompts) == 1 && prompt.Prompts[0].Role == "user"
		}),
		payload.Variables,
		"",
	).Return(mockResponse, nil)

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response openai.ChatCompletionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "Hello Alice! Nice to meet you.", response.Choices[0].Message.Content)
	assert.Equal(s.T(), 28, response.Usage.TotalTokens)

	s.iai.AssertExpectations(s.T())
}

func (s *promptTestSuite) TestTestPromptWithMultipleMessages() {
	// Test with system and user messages
	payload := testPromptPayload{
		ProjectID:  s.project.ID,
		ProviderID: s.provider.ID,
		Name:       "Test Prompt",
		Prompts: []schema.PromptRow{
			{
				Role:   "system",
				Prompt: "You are a helpful assistant.",
			},
			{
				Role:   "user",
				Prompt: "What is the capital of {{country}}?",
			},
		},
		Variables: map[string]string{
			"country": "France",
		},
	}

	mockResponse := openai.ChatCompletionResponse{
		ID:      "chatcmpl-789",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-3.5-turbo",
		Choices: []openai.ChatCompletionChoice{
			{
				Index: 0,
				Message: openai.ChatCompletionMessage{
					Role:    "assistant",
					Content: "The capital of France is Paris.",
				},
				FinishReason: "stop",
			},
		},
		Usage: openai.Usage{
			PromptTokens:     15,
			CompletionTokens: 8,
			TotalTokens:      23,
		},
	}

	s.iai.On("Chat",
		mock.AnythingOfType("context.backgroundCtx"),
		mock.MatchedBy(func(p *ent.Provider) bool {
			return p.ID == s.provider.ID && p.Name == "Test Provider"
		}),
		mock.MatchedBy(func(prompt ent.Prompt) bool {
			return len(prompt.Prompts) == 2 &&
				prompt.Prompts[0].Role == "system" &&
				prompt.Prompts[1].Role == "user"
		}),
		payload.Variables,
		"",
	).Return(mockResponse, nil)

	payloadBytes, err := json.Marshal(payload)
	assert.Nil(s.T(), err)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/test", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)

	testPrompt(c)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response openai.ChatCompletionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "The capital of France is Paris.", response.Choices[0].Message.Content)

	s.iai.AssertExpectations(s.T())
}

func (s *promptTestSuite) TearDownSuite() {
	service.EntClient.Provider.Delete().Where(provider.HasCreatorWith(user.ID(s.user.ID))).ExecX(context.Background())
	service.EntClient.Project.Delete().Where(project.HasCreatorWith(user.ID(s.user.ID))).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.user.ID).ExecX(context.Background())
	service.Close()
}

func TestPromptTestSuite(t *testing.T) {
	suite.Run(t, new(promptTestSuite))
}
