package schema

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent/prompt"
	dbSchema "github.com/PromptPal/PromptPal/ent/schema"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type promptTestSuite struct {
	suite.Suite
	pjID       int
	promptName string
	promptID   int
}

func (s *promptTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs, oi, w3)

	q := QueryResolver{}

	pjName := utils.RandStringRunes(1 << 4)
	openAIToken := utils.RandStringRunes(1 << 8)

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})
	pj, _ := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &pjName,
			OpenAIToken: &openAIToken,
		},
	})

	s.pjID = int(pj.ID())
	s.promptName = "test-prompt"
}

func (s *promptTestSuite) TestCreatePrompt() {
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
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
		},
	})

	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "test-prompt", result.Name())
	assert.Equal(s.T(), "test-prompt description", result.Description())
	assert.EqualValues(s.T(), 1, result.TokenCount())
	assert.NotEmpty(s.T(), result.ID())
	s.promptID = int(result.ID())
}

func (s *promptTestSuite) TestListPrompt() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
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

func (s *promptTestSuite) TestGetPrompt() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})
	result, err := q.Prompt(ctx, promptArgs{
		ID: int32(s.promptID),
	})
	assert.Nil(s.T(), err)
	pt := result
	assert.Equal(s.T(), "test-prompt", pt.Name())

	assert.NotEmpty(s.T(), pt.CreatedAt())
	assert.NotEmpty(s.T(), pt.UpdatedAt())
}

func (s *promptTestSuite) TestUpdatePrompt() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	truthy := true

	result, err := q.UpdatePrompt(ctx, updatePromptArgs{
		ID: int32(s.promptID),
		Data: createPromptData{
			ProjectID:   int32(s.pjID),
			Name:        "test-prompt-podcast-AsyncTalk",
			Description: "welcome to listen the podcast: `AsyncTalk`",
			TokenCount:  9231,
			Enabled:     &truthy,
			Debug:       &truthy,
			PublicLevel: prompt.PublicLevelPrivate,
			Prompts: []dbSchema.PromptRow{
				{
					Prompt: "AsyncTalk podcast is a a good chinese podcast talk about frontend development {{ var88 }}",
					Role:   "system",
				},
			},
			Variables: []dbSchema.PromptVariable{
				{
					Name: "var88",
					Type: "string",
				},
			},
		},
	})
	assert.Nil(s.T(), err)
	assert.True(s.T(), result.Debug())
	assert.True(s.T(), result.Enabled())
	assert.EqualValues(s.T(), "test-prompt", result.Name())
	assert.EqualValues(s.T(), 9231, result.TokenCount())
	assert.EqualValues(s.T(), s.promptID, result.ID())
	assert.EqualValues(s.T(), "private", result.PublicLevel())

	pts := result.Prompts()
	assert.Len(s.T(), pts, 1)
	pt := pts[0]
	assert.Equal(s.T(), "system", pt.Role())
	assert.Equal(s.T(), "AsyncTalk podcast is a a good chinese podcast talk about frontend development {{ var88 }}", pt.Prompt())

	vars := result.Variables()
	assert.Len(s.T(), vars, 1)
	var1 := vars[0]
	assert.Equal(s.T(), "var88", var1.Name())
	assert.Equal(s.T(), "string", var1.Type())

}

func (s *promptTestSuite) TearDownSuite() {
	service.Close()
}

func TestPromptTestSuite(t *testing.T) {
	suite.Run(t, new(promptTestSuite))
}
