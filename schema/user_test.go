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

type userTestSuite struct {
	suite.Suite
	pjID       int
	promptName string
	promptID   int
}

func (s *userTestSuite) SetupSuite() {
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
	pj, _ := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &pjName,
			OpenAIToken: &openAIToken,
		},
	})

	s.pjID = int(pj.ID())
	s.promptName = "test-prompt"
}

func (s *userTestSuite) TestGetUser() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	result, err := q.User(ctx, userArgs{
		ID: int32(theUserID),
	})

	assert.Nil(s.T(), err)

	assert.EqualValues(s.T(), theUserID, result.ID())
}

func (s *userTestSuite) TestGetMe() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	result, err := q.User(ctx, userArgs{
		ID: int32(-1),
	})

	assert.Nil(s.T(), err)

	assert.EqualValues(s.T(), theUserID, result.ID())
	assert.NotEmpty(s.T(), result.Name())
	assert.NotEmpty(s.T(), result.Addr())
}

func (s *userTestSuite) TearDownSuite() {
	service.EntClient.Project.DeleteOneID(s.pjID).ExecX(context.Background())
	service.Close()
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(userTestSuite))
}
