package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/project"
	"github.com/PromptPal/PromptPal/ent/prompt"
	"github.com/PromptPal/PromptPal/ent/promptcall"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/routes"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type callTestSuite struct {
	suite.Suite
	user         *ent.User
	pjID         int
	promptID     int
	providerID   int
	promptHashID string
	callID       int
	apiToken     string
	router       *gin.Engine
}

func (s *callTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewMockWeb3Service(s.T())
	iai := service.NewMockIsomorphicAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)
	Setup(hs, w3)

	// w3.
	// 	On(
	// 		"VerifySignature",
	// 		"0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
	// 		"message",
	// 		"signature",
	// 	).
	// 	Return(true, nil)

	q := QueryResolver{}

	// authRes, _ := q.Auth(context.Background(), authInput{
	// 	Auth: authAuthData{
	// 		Address:   "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
	// 		Signature: "signature",
	// 		Message:   "message",
	// 	},
	// })

	user, err := service.EntClient.User.
		Create().
		SetUsername("annatarhe_user_schema_call_test001").
		SetEmail("annatarhe_user_schema_call_test001@annatarhe.com").
		SetPasswordHash("hash").
		SetAddr("test-addr-annatarhe_user_schema_call_test001").
		SetName("Test User9").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())
	assert.Nil(s.T(), err)

	s.user = user

	pjName := "annatarhe_pj_schema_call_test"
	openAIToken := utils.RandStringRunes(1 << 8)

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: user.ID,
	})

	provider, err := q.CreateProvider(ctx, createProviderArgs{
		Data: createProviderData{
			Name:     pjName,
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   openAIToken,
			Config:   "{}",
		},
	})

	if err != nil {
		logrus.Println(err.Error())
	}
	assert.Nil(s.T(), err)

	pj, exp := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &pjName,
			OpenAIToken: &openAIToken,
			ProviderId:  provider.ID(),
		},
	})

	if exp != nil {
		logrus.Println(exp.Error())
	}
	assert.Nil(s.T(), exp)

	s.pjID = int(pj.ID())
	s.providerID = int(provider.ID())

	ot, _ := q.CreateOpenToken(ctx, createOpenTokenArgs{
		Data: createOpenTokenData{
			ProjectID:   int32(s.pjID),
			Name:        "test-openToken-call-test",
			Description: "open token for call test",
		},
	})

	s.apiToken = ot.Token()

	iai.On("GetProvider", mock.Anything, mock.Anything).
		Return(&ent.Provider{
			ID:       int(provider.ID()),
			Name:     provider.Name(),
			Source:   provider.Source(),
			Endpoint: provider.Endpoint(),
			ApiKey:   openAIToken,
			Config:   map[string]interface{}{},
		}, nil)

	iai.On(
		"Chat",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		map[string]string{
			"text": "var1",
		},
		"34",
	).
		After(time.Millisecond*100).
		Return(openai.ChatCompletionResponse{
			ID:      "j",
			Object:  "completion",
			Created: time.Now().Unix(),
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 1,
					Message: openai.ChatCompletionMessage{
						Content: "ji ni tai mei",
					},
					FinishReason: "completed",
				},
			},
			Usage: openai.Usage{
				PromptTokens:     18,
				CompletionTokens: 8888,
				TotalTokens:      1 << 16,
			},
		}, nil)

	s.router = routes.SetupGinRoutes("test", w3, iai, hs, nil)
}

func (s *callTestSuite) TestCreatePrompt() {
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.user.ID,
	})

	result, err := q.CreatePrompt(ctx, createPromptArgs{
		Data: createPromptData{
			ProjectID:   int32(s.pjID),
			Name:        "test-prompt",
			Description: "test-prompt description",
			TokenCount:  1,
			Debug:       nil,
			Enabled:     nil,
			Prompts: []dbSchema.PromptRow{
				{
					Prompt: "a-simple prompt {{ var1 }}",
					Role:   "system",
				},
			},
			Variables: []dbSchema.PromptVariable{
				{
					Name: "var1",
					Type: "string",
				},
			},
			PublicLevel: prompt.PublicLevelPublic,
			ProviderId:  int32(s.providerID),
		},
	})

	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "test-prompt", result.Name())
	assert.Equal(s.T(), "test-prompt description", result.Description())
	assert.EqualValues(s.T(), 1, result.TokenCount())
	assert.NotEmpty(s.T(), result.ID())
	s.promptID = int(result.ID())
	hid, _ := result.HashID()
	s.promptHashID = hid
}

func (s *callTestSuite) TestListPrompt() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.user.ID,
	})

	resp := q.Prompts(ctx, promptsArgs{
		ProjectID:  int32(s.pjID),
		Pagination: paginationInput{Limit: 20, Offset: 0},
	})

	cn, err := resp.Count(ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 1, cn)
	result, err := resp.Edges(ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), result, int(cn))

	pt := result[0]
	assert.Equal(s.T(), "test-prompt", pt.Name())
}

func (s *callTestSuite) TestGetPrompt() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.user.ID,
	})
	result, err := q.Prompt(ctx, promptArgs{
		ID: int32(s.promptID),
	})
	assert.Nil(s.T(), err)
	pt := result
	assert.Equal(s.T(), "test-prompt", pt.Name())
}
func (s *callTestSuite) TestPerformCall() {
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
		fmt.Sprintf("/api/v1/public/prompts/run/%s", s.promptHashID),
		strings.NewReader(payload),
	)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "API "+s.apiToken)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	var result struct {
		PromptID           string `json:"id"`
		ResponseMessage    string `json:"message"`
		ResponseTokenCount int    `json:"tokenCount"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.promptHashID, result.PromptID)
	assert.Equal(s.T(), "ji ni tai mei", result.ResponseMessage)
	assert.EqualValues(s.T(), 8888, result.ResponseTokenCount)

	// test project metrics
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.user.ID,
	})

	pj, err := q.Project(ctx, projectArgs{
		ID: int32(s.pjID),
	})
	assert.Nil(s.T(), err)

	rcs, err := pj.PromptMetrics().RecentCounts(ctx)
	assert.Nil(s.T(), err)

	assert.Len(s.T(), rcs, 1)

	rc := rcs[0]
	assert.EqualValues(s.T(), 1, rc.Count())
	_hid, err := rc.Prompt().HashID()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), s.promptHashID, _hid)
	lastCalls := rc.Prompt().LatestCalls(ctx)
	cc, err := lastCalls.Count(ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 1, cc)
	edges, err := lastCalls.Edges(ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), edges, 1)
	edge := edges[0]

	assert.GreaterOrEqual(s.T(), edge.ID(), int32(1))
	// because the debug is disabled
	// assert.EqualValues(s.T(), "ji ni tai mei", edge.Message())
	assert.Nil(s.T(), edge.Message())
	assert.GreaterOrEqual(s.T(), edge.Duration(), int32(100))
	assert.EqualValues(s.T(), "34", edge.UserId())
	assert.EqualValues(s.T(), 8888, edge.ResponseToken())
	assert.EqualValues(s.T(), "success", edge.Result())
	assert.NotEmpty(s.T(), edge.CreatedAt())

	cs := q.Calls(ctx, callsArgs{
		PromptID: int32(s.promptID),
		Pagination: paginationInput{
			Limit:  10,
			Offset: 0,
		},
	})

	cCount, err := cs.Count(ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 1, cCount)

	edges, err = cs.Edges(ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), edges, 1)
}

func (s *callTestSuite) TearDownSuite() {
	service.EntClient.PromptCall.Delete().Where(promptcall.HasProjectWith(project.ID(s.pjID))).ExecX(context.Background())
	service.EntClient.Prompt.DeleteOneID(s.promptID).ExecX(context.Background())
	service.EntClient.Project.DeleteOneID(s.pjID).ExecX(context.Background())
	service.EntClient.Provider.DeleteOneID(s.providerID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.user.ID).ExecX(context.Background())
	service.Close()
}

func TestCallTestSuite(t *testing.T) {
	suite.Run(t, new(callTestSuite))
}
