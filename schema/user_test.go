package schema

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type userTestSuite struct {
	suite.Suite
	user       *ent.User
	pjID       int
	promptName string
	promptID   int
	providerID int
}

func (s *userTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewMockWeb3Service(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	service.InitRedis(config.GetRuntimeConfig().RedisURL)

	Setup(hs, w3)

	q := QueryResolver{}

	pjName := "annatarhe_pj_schema_user_test"
	openAIToken := utils.RandStringRunes(1 << 8)

	user, err := service.EntClient.User.
		Create().
		SetUsername("annatarhe_user_schema_user_test").
		SetEmail("annatarhe_user_schema_user_test001@annatarhe.com").
		SetPasswordHash("hash").
		SetAddr("test-addr-annatarhe_user_schema_user_test001").
		SetName("Test User9").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())
	assert.Nil(s.T(), err)

	s.user = user

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: user.ID,
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
	pj, exp := q.CreateProject(ctx, createProjectArgs{
		Data: createProjectData{
			Name:        &pjName,
			OpenAIToken: &openAIToken,
			ProviderId:  int32(provider.ID()),
		},
	})

	assert.Nil(s.T(), exp)

	s.pjID = int(pj.ID())
	s.promptName = "test-prompt"
	s.providerID = int(provider.ID())
}

func (s *userTestSuite) TestGetUser() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	id := int32(theUserID)
	result, err := q.User(ctx, userArgs{
		ID: &id,
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
		ID: nil,
	})

	assert.Nil(s.T(), err)

	assert.EqualValues(s.T(), theUserID, result.ID())
	assert.NotEmpty(s.T(), result.Name())
	assert.NotEmpty(s.T(), result.Addr())
}

func (s *userTestSuite) TestGetUserWithAllFields() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	id := int32(theUserID)
	result, err := q.User(ctx, userArgs{
		ID: &id,
	})

	assert.Nil(s.T(), err)

	// Test all fields are present
	assert.EqualValues(s.T(), theUserID, result.ID())
	assert.NotEmpty(s.T(), result.Name())
	assert.NotEmpty(s.T(), result.Addr())

	// Test new fields (may be empty but should not panic)
	assert.NotNil(s.T(), result.Avatar())
	assert.NotNil(s.T(), result.Email())
	assert.NotNil(s.T(), result.Phone())
	assert.NotNil(s.T(), result.Lang())
	assert.GreaterOrEqual(s.T(), result.Level(), int32(0))
	assert.NotNil(s.T(), result.Source())
}

func (s *userTestSuite) TestGetCurrentUserWithAllFields() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	// Test with nil ID (current user)
	result, err := q.User(ctx, userArgs{
		ID: nil,
	})

	assert.Nil(s.T(), err)

	// Test all fields are present for current user
	assert.EqualValues(s.T(), theUserID, result.ID())
	assert.NotEmpty(s.T(), result.Name())
	assert.NotEmpty(s.T(), result.Addr())

	// Test new fields
	assert.NotNil(s.T(), result.Avatar())
	assert.NotNil(s.T(), result.Email())
	assert.NotNil(s.T(), result.Phone())
	assert.NotNil(s.T(), result.Lang())
	assert.GreaterOrEqual(s.T(), result.Level(), int32(0))
	assert.NotNil(s.T(), result.Source())
}

func (s *userTestSuite) TestGetUserNotFound() {
	theUserID := service.EntClient.User.Query().FirstIDX(context.Background())
	q := QueryResolver{}

	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: theUserID,
	})

	// Test with non-existent user ID
	nonExistentID := int32(99999)
	result, err := q.User(ctx, userArgs{
		ID: &nonExistentID,
	})

	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), result.u)
}

func (s *userTestSuite) TearDownSuite() {
	service.EntClient.Project.DeleteOneID(s.pjID).ExecX(context.Background())
	service.EntClient.Provider.DeleteOneID(s.providerID).ExecX(context.Background())
	service.EntClient.User.DeleteOneID(s.user.ID).ExecX(context.Background())
	service.Close()
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(userTestSuite))
}
