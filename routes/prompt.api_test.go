package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/otiai10/openaigo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type promptPublicAPITestSuite struct {
	suite.Suite
	oi        *mocks.OpenAIService
	router    *gin.Engine
	token     string
	apiToken  string
	pjName    string
	promptHid string
	promptId  int
}

func (s *promptPublicAPITestSuite) SetupTest() {
	// just for get a token
	u := new(userTestSuite)
	u.SetT(s.T())
	u.SetS(&s.Suite)
	u.SetupTest()
	authInfo, _ := u.GetAuthToken()
	s.token = authInfo.Token
	u.TearDownSuite()

	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	s.oi = oi

	service.InitDB()
	s.router = SetupGinRoutes("test", w3, oi, hs)
}

func (s *promptPublicAPITestSuite) TestCreateProjectAndPrompt() {
	// 1. create project
	w := httptest.NewRecorder()
	s.pjName = RandStringRunes(1 << 6)
	payload := fmt.Sprintf(`{"name": "%s", "openaiToken": "openaiToken"}`, s.pjName)

	req, _ := http.NewRequest("POST", "/api/v1/admin/projects", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ent.Project{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.pjName, result.Name)
	assert.Equal(s.T(), "https://api.openai.com/v1", result.OpenAIBaseURL)
	assert.NotEmpty(s.T(), result.ID)

	// 2. create prompt
	w2 := httptest.NewRecorder()
	promptName := RandStringRunes(1 << 5)
	payload2 := fmt.Sprintf(`
	{
		"projectId": %d,
		"name": "%s",
		"description": "description",
		"tokenCount": 17,
		"prompts": [
			{
				"prompt": "prompt {{ text }}",
				"role": "system"
			}
		],
		"variables": [
			{
				"name": "text",
				"type": "string"
			}
		],
		"publicLevel": "public"
	}
	`, result.ID, promptName)

	req2, _ := http.NewRequest("POST", "/api/v1/admin/prompts", strings.NewReader(payload2))
	req2.Header.Add("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w2, req2)
	assert.Equal(s.T(), 200, w2.Code)

	result2 := internalPromptItem{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &result2)
	assert.Nil(s.T(), err2)

	assert.Equal(s.T(), promptName, result2.Prompt.Name)
	assert.NotEmpty(s.T(), result2.Prompt.ID)
	assert.NotEmpty(s.T(), result2.HashID)
	s.promptHid = result2.HashID
	s.promptId = result2.Prompt.ID

	// 3. get api token
	w3 := httptest.NewRecorder()
	apiTokenName := RandStringRunes(1 << 4)
	payload3 := fmt.Sprintf(`
	{
		"name": "%s",
		"description": "apiToken",
		"ttl": 10000
	}
	`, apiTokenName)

	req3, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/admin/projects/%d/open-tokens", result.ID),
		strings.NewReader(payload3),
	)
	req3.Header.Add("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w3, req3)
	assert.Equal(s.T(), 200, w3.Code)

	result3 := map[string]string{}
	err3 := json.Unmarshal(w3.Body.Bytes(), &result3)
	assert.Nil(s.T(), err3)

	assert.NotEmpty(s.T(), result3["token"])
	s.apiToken = result3["token"]
}

func (s *promptPublicAPITestSuite) TestPublicAPIListProject() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/public/prompts?limit=10&cursor=9999999", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "API "+s.apiToken)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ListResponse[publicPromptItem]{}

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), 1, result.Count)
	assert.Equal(s.T(), 1, len(result.Data))

	d := result.Data[0]

	assert.Equal(s.T(), s.promptHid, d.HashID)
	assert.Equal(s.T(), 17, d.TokenCount)
}

func (s *promptPublicAPITestSuite) TestPublicAPIRunPrompt() {
	s.oi.On(
		"Chat",
		mock.Anything,
		mock.Anything,
		[]schema.PromptRow{
			{
				Prompt: "prompt {{ text }}",
				Role:   "system",
			},
		},
		map[string]string{
			"text": "var1",
		},
		"34",
	).Return(openaigo.ChatCompletionResponse{
		ID:      "j",
		Object:  "completion",
		Created: time.Now().Unix(),
		Choices: []openaigo.Choice{
			{
				Index: 1,
				Message: openaigo.Message{
					Content: "ji ni tai mei",
				},
				FinishReason: "completed",
			},
		},
		Usage: openaigo.Usage{
			PromptTokens:     18,
			CompletionTokens: 8888,
			TotalTokens:      1 << 16,
		},
	}, nil)

	w := httptest.NewRecorder()
	payload := `
	{
		"variables": {
			"text": "var1"
		},
		"userId": "34"
	}
	`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("/api/v1/public/prompts/run/%s", s.promptHid),
		strings.NewReader(payload),
	)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "API "+s.apiToken)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := apiRunPromptResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.promptHid, result.PromptID)

	assert.Equal(s.T(), 8888, result.ResponseTokenCount)
	assert.Equal(s.T(), "ji ni tai mei", result.ResponseMessage)
}

func (s *promptPublicAPITestSuite) TearViewPromptCalls() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("/api/v1/admin/prompts/%d/calls?limit=10&cursor=9999999", s.promptId),
		nil,
	)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "API "+s.apiToken)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ListResponse[*ent.PromptCall]{}

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), 1, result.Count)
	assert.Equal(s.T(), 1, len(result.Data))
	d := result.Data[0]
	assert.Equal(s.T(), 1<<16, d.TotalToken)
	assert.Greater(s.T(), d.Duration, 0)
	assert.Equal(s.T(), nil, d.Message)
}

func (s *promptPublicAPITestSuite) TearDownSuite() {
	service.Close()
}

func TestPromptPublicAPITestSuite(t *testing.T) {
	suite.Run(t, new(promptPublicAPITestSuite))
}
