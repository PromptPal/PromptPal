package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/provider"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type promptAPITestSuite struct {
	suite.Suite
	router   *gin.Engine
	w3       *service.MockWeb3Service
	iai      *service.MockIsomorphicAIService
	hashid   *service.MockHashIDService
	user     *ent.User
	project  *ent.Project
	prompt   *ent.Prompt
	provider *ent.Provider
}

func (s *promptAPITestSuite) SetupTest() {
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
	s.router = SetupGinRoutes("test", s.w3, s.iai, s.hashid, nil)

	// Create test data
	s.createTestData()
}

func (s *promptAPITestSuite) createTestData() {
	// Create test user
	user, err := service.EntClient.User.
		Create().
		SetUsername("annnatarhe.route_prompt_api").
		SetEmail("annnatarhe.route_prompt_api@example.com").
		SetAddr("0x1234567890" + "route_prompt_api").
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

	// Create test prompt
	promptRows := []schema.PromptRow{
		{
			Role:   "user",
			Prompt: "Hello {{name}}",
		},
	}
	variables := []schema.PromptVariable{
		{
			Name: "name",
			Type: "string",
		},
	}

	prompt, err := service.EntClient.Prompt.
		Create().
		SetName("Test Prompt").
		SetDescription("Test prompt description").
		SetProjectId(project.ID).
		SetProviderId(provider.ID).
		SetPrompts(promptRows).
		SetCreatorID(user.ID).
		SetVariables(variables).
		SetTokenCount(10).
		SetDebug(true).
		Save(context.Background())
	assert.Nil(s.T(), err)
	s.prompt = prompt
}

func (s *promptAPITestSuite) getAuthHeaders() map[string]string {
	// Mock JWT auth - in real scenarios would use actual JWT
	// For testing, we'll rely on test mode setup
	return map[string]string{
		"Authorization": "Bearer test-token",
	}
}

func (s *promptAPITestSuite) TestAPIListPrompts() {
	// Mock hashid encode for prompt ID
	s.hashid.On("Encode", s.prompt.ID).Return("abc123", nil)

	// Set up context with project ID
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/prompts?limit=10&cursor=1000", nil)

	// Add headers
	for k, v := range s.getAuthHeaders() {
		req.Header.Set(k, v)
	}

	// Create a context with authenticated user and project
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)
	c.Set("pid", s.project.ID)

	// Call the handler
	apiListPrompts(c)

	// Assert response
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response ListResponse[publicPromptItem]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.GreaterOrEqual(s.T(), response.Count, 1)
	assert.Len(s.T(), response.Data, 1)
	assert.Equal(s.T(), "Test Prompt", response.Data[0].Name)
}

func (s *promptAPITestSuite) TestAPIListPromptsInvalidQuery() {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/prompts?limit=invalid", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("uid", s.user.ID)
	c.Set("pid", s.project.ID)

	apiListPrompts(c)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *promptAPITestSuite) TestAPIRunPromptMiddleware() {
	// Mock hashid service
	hashedID := "abc123"
	s.hashid.On("Decode", hashedID).Return(s.prompt.ID, nil)

	gin.SetMode(gin.TestMode)

	payload := apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	}
	payloadBytes, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/run", hashedID), bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("pid", s.project.ID)

	// Call middleware
	apiRunPromptMiddleware(c)

	// Should not abort if successful
	assert.False(s.T(), c.IsAborted())

	// Check context values
	promptData, exists := c.Get("prompt")
	assert.True(s.T(), exists)
	prompt := promptData.(ent.Prompt)
	assert.Equal(s.T(), s.prompt.ID, prompt.ID)
}

func (s *promptAPITestSuite) TestAPIRunPromptMiddlewareInvalidID() {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/prompts/invalid/run", nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	// Mock hashid decode failure
	s.hashid.On("Decode", "invalid").Return(0, fmt.Errorf("invalid hash"))

	apiRunPromptMiddleware(c)

	assert.True(s.T(), c.IsAborted())
	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *promptAPITestSuite) TestAPIRunPrompt() {
	// Set up mocks
	hashedID := "abc123"

	// Mock IsomorphicAIService calls
	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	mockResponse := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: "Hello John",
				},
			},
		},
		Usage: openai.Usage{
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	s.iai.On("Chat", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(p *ent.Provider) bool {
		return p.ID == s.provider.ID
	}), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	}), map[string]string{"name": "John"}, "user123").Return(mockResponse, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/run", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	apiRunPrompt(c)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response service.APIRunPromptResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), hashedID, response.PromptID)
	assert.Equal(s.T(), "Hello John", response.ResponseMessage)
	assert.Equal(s.T(), 5, response.ResponseTokenCount)
}

func (s *promptAPITestSuite) TestAPIRunPromptGetProviderError() {
	hashedID := "abc123"

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(nil, fmt.Errorf("provider error"))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/run", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	apiRunPrompt(c)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *promptAPITestSuite) TestAPIRunPromptChatError() {
	hashedID := "abc123"

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	s.iai.On("Chat", mock.Anything, s.provider, mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	}), map[string]string{"name": "John"}, "user123").Return(openai.ChatCompletionResponse{}, fmt.Errorf("chat error"))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/run", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	apiRunPrompt(c)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *promptAPITestSuite) TestAPIRunPromptNoChoices() {
	hashedID := "abc123"
	mockResponse := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{}, // Empty choices
		Usage: openai.Usage{
			CompletionTokens: 0,
			TotalTokens:      10,
		},
	}

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	s.iai.On("Chat", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(p *ent.Provider) bool {
		return p.ID == s.provider.ID
	}), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	}), map[string]string{"name": "John"}, "user123").Return(mockResponse, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/run", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	apiRunPrompt(c)

	assert.Equal(s.T(), http.StatusBadRequest, w.Code)
}

func (s *promptAPITestSuite) TestAPIRunPromptStream() {
	hashedID := "abc123"

	// Create mock stream response
	mockStreamResponse := &service.ChatStreamResponse{
		Done:    make(chan bool, 1),
		Err:     make(chan error, 1),
		Info:    make(chan openai.Usage, 1),
		Message: make(chan []openai.ChatCompletionChoice, 1),
	}

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	s.iai.On("ChatStream", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(p *ent.Provider) bool {
		return p.ID == s.provider.ID
	}), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	}), map[string]string{"name": "John"}, "user123").Return(mockStreamResponse, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/stream", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	// Send test data to channels
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockStreamResponse.Message <- []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: "Hello",
				},
			},
		}
		time.Sleep(10 * time.Millisecond)
		mockStreamResponse.Info <- openai.Usage{CompletionTokens: 1}
		time.Sleep(10 * time.Millisecond)
		mockStreamResponse.Done <- true
	}()

	apiRunPromptStream(c)

	// Verify headers are set for streaming
	assert.Equal(s.T(), "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(s.T(), "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(s.T(), "keep-alive", w.Header().Get("Connection"))
}

func (s *promptAPITestSuite) TestAPIRunPromptStreamError() {
	hashedID := "abc123"

	s.iai.On("GetProvider", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	})).Return(s.provider, nil)

	s.iai.On("ChatStream", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(p *ent.Provider) bool {
		return p.ID == s.provider.ID
	}), mock.MatchedBy(func(prompt ent.Prompt) bool {
		return prompt.ID == s.prompt.ID
	}), map[string]string{"name": "John"}, "user123").Return(nil, fmt.Errorf("stream error"))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/prompts/%s/stream", hashedID), nil)

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: hashedID}}
	c.Set("prompt", *s.prompt)
	c.Set("pj", *s.project)
	c.Set("payload", apiRunPromptPayload{
		Variables: map[string]string{"name": "John"},
		UserId:    "user123",
	})

	apiRunPromptStream(c)

	assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *promptAPITestSuite) TearDownSuite() {
	service.EntClient.Prompt.Delete().Where(prompt.HasCreatorWith(user.ID(s.user.ID))).ExecX(context.Background())
	service.EntClient.Provider.Delete().Where(provider.HasCreatorWith(user.ID(s.user.ID))).ExecX(context.Background())
	service.EntClient.Project.Delete().Where(project.HasCreatorWith(user.ID(s.user.ID))).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.user.ID).ExecX(context.Background())
	service.Close()
}

func TestPromptAPITestSuite(t *testing.T) {
	suite.Run(t, new(promptAPITestSuite))
}
