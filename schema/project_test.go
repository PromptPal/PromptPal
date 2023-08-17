package schema

import (
	"context"
	"errors"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type projectTestSuite struct {
	suite.Suite
	projectName string
	projectID   int
}

func (s *projectTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs, oi, w3)

	s.projectName = "test-project"
}

func (s *projectTestSuite) TestCreateProject() {
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	tk := "SOME_RANDOM_TOKEN_HERE"

	result, err := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &s.projectName,
			OpenAIToken: &tk,
		},
	})

	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.projectName, result.Name())
	assert.Equal(s.T(), "https://api.openai.com/v1", result.OpenAIBaseURL())
	assert.NotEmpty(s.T(), result.ID())
	s.projectID = int(result.ID())
}

func (s *projectTestSuite) TestListProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})
	result, err := q.Projects(ctx, projectsArgs{
		Pagination: paginationInput{Limit: 20, Offset: 0},
	})
	assert.Nil(s.T(), err)
	edges := result.Edges()
	assert.EqualValues(s.T(), 1, result.Count())
	assert.Len(s.T(), edges, int(result.Count()))

	pj := edges[0]
	assert.Equal(s.T(), s.projectName, pj.Name())
}

func (s *projectTestSuite) TestGetProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})
	result, err := q.Project(ctx, projectArgs{
		ID: int32(s.projectID),
	})
	assert.Nil(s.T(), err)
	pj := result
	assert.Equal(s.T(), s.projectName, pj.Name())
	assert.Equal(s.T(), "https://api.openai.com/v1", pj.OpenAIBaseURL())
	assert.Equal(s.T(), "gpt-3.5-turbo", pj.OpenAIModel())
	assert.Equal(s.T(), "SOME_RANDOM_TOKEN_HERE", pj.OpenAIToken())
	assert.EqualValues(s.T(), 1, pj.OpenAITemperature())
	assert.EqualValues(s.T(), 0.9, pj.OpenAITopP())
	assert.EqualValues(s.T(), 0, pj.OpenAIMaxTokens())
	assert.NotEmpty(s.T(), pj.CreatedAt())
	assert.NotEmpty(s.T(), pj.UpdatedAt())

	creator, err := pj.Creator(ctx)
	assert.Nil(s.T(), err)

	assert.EqualValues(s.T(), 1, creator.ID())

	lps := pj.LatestPrompts(ctx)
	cs, err := lps.Count(ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), cs, 0)
	lps2, err := lps.Edges(ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), lps2, 0)

	_, err = q.Project(ctx, projectArgs{
		ID: int32(887771),
	})
	assert.Error(s.T(), err)

	ge, ok := err.(GraphQLHttpError)
	assert.True(s.T(), ok)
	assert.EqualValues(s.T(), "[500]: ent: project not found", ge.Error())
	assert.EqualValues(s.T(), errors.New("[500]: ent: project not found"), ge.Unwrap())
}

func (s *projectTestSuite) TestUpdateProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	openAiMaxTokens := 78
	truthy := true
	pj, err := q.UpdateProject(ctx, updateProjectArgs{
		ID: int32(s.projectID),
		Data: createProjectData{
			Enabled:         &truthy,
			OpenAIMaxTokens: &openAiMaxTokens,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.projectName, pj.Name())
	assert.EqualValues(s.T(), openAiMaxTokens, pj.OpenAIMaxTokens())
	assert.True(s.T(), pj.Enabled())

	r, err := q.DeleteProject(ctx, deleteProjectArgs{
		ID: int32(s.projectID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), r)
}

func (s *projectTestSuite) TearDownSuite() {
	service.Close()
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(projectTestSuite))
}
