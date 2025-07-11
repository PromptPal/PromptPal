package schema

import (
	"context"
	"log"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type projectTestSuite struct {
	suite.Suite
	uid         int
	projectName string
	projectID   int
	providerID  int
}

func (s *projectTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewMockWeb3Service(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)
	Setup(hs, w3)

	u := service.
		EntClient.
		User.
		Create().
		SetAddr(utils.RandStringRunes(1 << 4)).
		SetName(utils.RandStringRunes(1 << 4)).
		SetLang("en").
		SetPhone(utils.RandStringRunes(1 << 4)).
		SetLevel(255).
		SetEmail(utils.RandStringRunes(1 << 3)).
		SaveX(context.Background())
	s.uid = u.ID
	s.projectName = utils.RandStringRunes(1 << 4)
}

func (s *projectTestSuite) TestCreateProject() {
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})

	tk := "SOME_RANDOM_TOKEN_HERE"

	provider, _ := q.CreateProvider(ctx, createProviderArgs{
		Data: createProviderData{
			Name:     s.projectName,
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   tk,
			Config:   "{}",
		},
	})

	result, err := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &s.projectName,
			OpenAIToken: &tk,
			ProviderId:  int32(provider.ID()),
		},
	})

	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.projectName, result.Name())
	assert.Equal(s.T(), "https://api.openai.com", result.OpenAIBaseURL())
	assert.Equal(s.T(), "https://generativelanguage.googleapis.com", result.GeminiBaseURL())
	assert.NotEmpty(s.T(), result.ID())
	s.projectID = int(result.ID())
	s.providerID = int(provider.ID())
}

func (s *projectTestSuite) TestListProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})
	result, err := q.Projects(ctx, projectsArgs{
		Pagination: paginationInput{Limit: 20, Offset: 0},
	})
	assert.Nil(s.T(), err)
	edges := result.Edges()
	log.Println("result: ", result.projects)
	assert.EqualValues(s.T(), 1, result.Count())
	assert.Len(s.T(), edges, int(result.Count()))

	pj := edges[0]
	assert.Equal(s.T(), s.projectName, pj.Name())
}

func (s *projectTestSuite) TestGetProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})
	result, err := q.Project(ctx, projectArgs{
		ID: int32(s.projectID),
	})
	assert.Nil(s.T(), err)
	pj := result
	assert.Equal(s.T(), s.projectName, pj.Name())
	assert.Equal(s.T(), "https://api.openai.com", pj.OpenAIBaseURL())
	assert.Equal(s.T(), "https://generativelanguage.googleapis.com", result.GeminiBaseURL())
	assert.Equal(s.T(), "gpt-3.5-turbo", pj.OpenAIModel())
	assert.Equal(s.T(), "SOME_RANDOM_TOKEN_HERE", pj.OpenAIToken())
	assert.EqualValues(s.T(), 1, pj.OpenAITemperature())
	assert.EqualValues(s.T(), 0.9, pj.OpenAITopP())
	assert.EqualValues(s.T(), 0, pj.OpenAIMaxTokens())
	assert.NotEmpty(s.T(), pj.CreatedAt())
	assert.NotEmpty(s.T(), pj.UpdatedAt())

	creator, err := pj.Creator(ctx)
	assert.Nil(s.T(), err)

	assert.EqualValues(s.T(), s.uid, creator.ID())

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
	assert.EqualValues(s.T(),
		map[string]interface{}{
			"code":    500,
			"message": "ent: project not found",
		},
		ge.Extensions(),
	)
}

func (s *projectTestSuite) TestUpdateProject() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.uid,
	})

	openAiMaxTokens := int32(78)
	temperature := float64(7.888888888)
	burl := "https://api.openai.com/v8"
	model := "annatarhe-35-turbo"
	truthy := true
	pj, err := q.UpdateProject(ctx, updateProjectArgs{
		ID: int32(s.projectID),
		Data: createProjectData{
			Enabled:           &truthy,
			OpenAIBaseURL:     &burl,
			OpenAIModel:       &model,
			OpenAIToken:       &s.projectName,
			OpenAITemperature: &temperature,
			OpenAIMaxTokens:   &openAiMaxTokens,
			OpenAITopP:        &temperature,
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.projectName, pj.Name())
	assert.EqualValues(s.T(), openAiMaxTokens, pj.OpenAIMaxTokens())
	assert.EqualValues(s.T(), temperature, pj.OpenAITemperature())
	assert.EqualValues(s.T(), burl, pj.OpenAIBaseURL())
	assert.EqualValues(s.T(), model, pj.OpenAIModel())
	assert.EqualValues(s.T(), s.projectName, pj.OpenAIToken())
	assert.True(s.T(), pj.Enabled())

	r, err := q.DeleteProject(ctx, deleteProjectArgs{
		ID: int32(s.projectID),
	})

	r2, err2 := q.DeleteProvider(ctx, deleteProviderArgs{
		ID: int32(s.providerID),
	})

	assert.Nil(s.T(), err)
	assert.True(s.T(), r)
	assert.Nil(s.T(), err2)
	assert.True(s.T(), r2)
}

func (s *projectTestSuite) TearDownSuite() {
	service.Close()
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(projectTestSuite))
}
