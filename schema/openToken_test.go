package schema

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type openTokenTestSuite struct {
	suite.Suite
	pjID       int
	providerID int
	otID       int
}

func (s *openTokenTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs, w3)

	q := QueryResolver{}

	pjName := utils.RandStringRunes(1 << 4)
	openAIToken := utils.RandStringRunes(1 << 8)

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	provider, _ := q.CreateProvider(ctx, createProviderArgs{
		Data: createProviderData{
			Name:     pjName,
			Source:   "openai",
			Endpoint: "https://api.openai.com/v1",
			ApiKey:   openAIToken,
			Config:   "{}",
		},
	})
	pj, _ := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &pjName,
			OpenAIToken: &openAIToken,
			ProviderId:  int32(provider.ID()),
		},
	})

	s.pjID = int(pj.ID())
	s.providerID = int(provider.ID())
}

func (s *openTokenTestSuite) TestCreateOpenToken() {
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	name := utils.RandStringRunes(1 << 3)
	desc := utils.RandStringRunes(1 << 4)

	result, err := q.CreateOpenToken(ctx, createOpenTokenArgs{
		Data: createOpenTokenData{
			ProjectID:   int32(s.pjID),
			Name:        name,
			Description: desc,
			TTL:         111,
		},
	})

	assert.Nil(s.T(), err)

	assert.NotEmpty(s.T(), result.Token())
	assert.GreaterOrEqual(s.T(), result.Data().ID(), int32(1))
	assert.Equal(s.T(), name, result.Data().Name())
	assert.Equal(s.T(), desc, result.Data().Description())
	assert.NotEmpty(s.T(), 1, result.Data().ExpireAt())

	s.otID = int(result.Data().ID())
}

func (s *openTokenTestSuite) TestListOpenToken() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	resp, err := q.Project(ctx, projectArgs{
		ID: int32(s.pjID),
	})

	assert.Nil(s.T(), err)
	ots, err := resp.OpenTokens(ctx)
	assert.Nil(s.T(), err)

	cn, err := ots.Count(ctx)
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 1, cn)
	result, err := ots.Edges(ctx)
	assert.Nil(s.T(), err)
	assert.Len(s.T(), result, int(cn))

	pt := result[0]
	assert.NotEmpty(s.T(), pt.Name())
}

func (s *openTokenTestSuite) TestPurgeOpenToken() {
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: 1,
	})

	result, err := q.DeleteOpenToken(ctx, deleteOpenTokenArgs{
		ID: int32(s.otID),
	})
	assert.Nil(s.T(), err)
	assert.True(s.T(), result)
}

func (s *openTokenTestSuite) TearDownSuite() {
	service.EntClient.Project.DeleteOneID(s.pjID).ExecX(context.Background())
	service.EntClient.Provider.DeleteOneID(s.providerID).ExecX(context.Background())
	service.Close()
}

func TestOpenTokenTestSuite(t *testing.T) {
	suite.Run(t, new(openTokenTestSuite))
}
