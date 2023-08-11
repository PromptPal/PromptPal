package schema

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
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
	// w3 := mocks.NewWeb3Service(s.T())
	// oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs)

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
}

func (s *projectTestSuite) TearDownSuite() {
	service.Close()
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(projectTestSuite))
}
