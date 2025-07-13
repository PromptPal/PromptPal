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

func (s *userTestSuite) TestCreateUser() {
	// Create an admin user first
	adminUser, err := service.EntClient.User.
		Create().
		SetUsername("admin_test_user").
		SetEmail("admin@test.com").
		SetPasswordHash("hash").
		SetAddr("admin-addr").
		SetName("Admin User").
		SetPhone("").
		SetLang("en").
		SetLevel(200). // Admin level > 100
		Save(context.Background())
	assert.Nil(s.T(), err)
	defer service.EntClient.User.DeleteOneID(adminUser.ID).ExecX(context.Background())

	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: adminUser.ID,
	})

	// Test successful user creation
	result, err := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.User().u)
	assert.Equal(s.T(), "Test User", result.User().Name())
	assert.Equal(s.T(), "test@example.com", result.User().Email())
	assert.NotEmpty(s.T(), result.Password()) // Should have generated password
	assert.Equal(s.T(), int32(1), result.User().Level()) // Default level
	assert.Equal(s.T(), "password", result.User().Source())

	// Clean up created user
	defer service.EntClient.User.DeleteOneID(int(result.User().ID())).ExecX(context.Background())
}

func (s *userTestSuite) TestCreateUserWithAllFields() {
	// Create an admin user first
	adminUser, err := service.EntClient.User.
		Create().
		SetUsername("admin_test_user2").
		SetEmail("admin2@test.com").
		SetPasswordHash("hash").
		SetAddr("admin-addr2").
		SetName("Admin User 2").
		SetPhone("").
		SetLang("en").
		SetLevel(255). // Max admin level
		Save(context.Background())
	assert.Nil(s.T(), err)
	defer service.EntClient.User.DeleteOneID(adminUser.ID).ExecX(context.Background())

	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: adminUser.ID,
	})

	phone := "123-456-7890"
	lang := "fr"
	level := int32(50)
	avatar := "https://example.com/avatar.jpg"
	username := "testuser123"

	// Test user creation with all optional fields
	result, err := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:     "Full Test User",
			Email:    "fulltest@example.com",
			Phone:    &phone,
			Lang:     &lang,
			Level:    &level,
			Avatar:   &avatar,
			Username: &username,
		},
	})

	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), result.User().u)
	assert.Equal(s.T(), "Full Test User", result.User().Name())
	assert.Equal(s.T(), "fulltest@example.com", result.User().Email())
	assert.Equal(s.T(), phone, result.User().Phone())
	assert.Equal(s.T(), lang, result.User().Lang())
	assert.Equal(s.T(), level, result.User().Level())
	assert.Equal(s.T(), avatar, result.User().Avatar())
	assert.NotEmpty(s.T(), result.Password())

	// Clean up created user
	defer service.EntClient.User.DeleteOneID(int(result.User().ID())).ExecX(context.Background())
}

func (s *userTestSuite) TestCreateUserNonAdmin() {
	// Use regular user (level 1)
	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: s.user.ID, // Regular user with level 1
	})

	// Test should fail for non-admin user
	result, err := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:  "Should Fail",
			Email: "fail@example.com",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "admin privileges required")
	assert.Empty(s.T(), result.u)
}

func (s *userTestSuite) TestCreateUserUnauthenticated() {
	q := QueryResolver{}
	// No context or invalid context
	ctx := context.Background()

	// Test should fail without authentication
	result, err := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:  "Should Fail",
			Email: "fail@example.com",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "authentication required")
	assert.Empty(s.T(), result.u)
}

func (s *userTestSuite) TestCreateUserPasswordGeneration() {
	// Create an admin user first
	adminUser, err := service.EntClient.User.
		Create().
		SetUsername("admin_test_user3").
		SetEmail("admin3@test.com").
		SetPasswordHash("hash").
		SetAddr("admin-addr3").
		SetName("Admin User 3").
		SetPhone("").
		SetLang("en").
		SetLevel(150). // Admin level > 100
		Save(context.Background())
	assert.Nil(s.T(), err)
	defer service.EntClient.User.DeleteOneID(adminUser.ID).ExecX(context.Background())

	q := QueryResolver{}
	ctx := context.WithValue(context.Background(), service.GinGraphQLContextKey, service.GinGraphQLContextType{
		UserID: adminUser.ID,
	})

	// Create multiple users and verify passwords are different
	result1, err1 := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:  "Test User 1",
			Email: "test1@example.com",
		},
	})
	assert.Nil(s.T(), err1)
	defer service.EntClient.User.DeleteOneID(int(result1.User().ID())).ExecX(context.Background())

	result2, err2 := q.CreateUser(ctx, createUserArgs{
		Data: createUserData{
			Name:  "Test User 2",
			Email: "test2@example.com",
		},
	})
	assert.Nil(s.T(), err2)
	defer service.EntClient.User.DeleteOneID(int(result2.User().ID())).ExecX(context.Background())

	// Passwords should be different
	assert.NotEqual(s.T(), result1.Password(), result2.Password())
	
	// Passwords should be of expected length (12 characters)
	assert.Equal(s.T(), 12, len(result1.Password()))
	assert.Equal(s.T(), 12, len(result2.Password()))

	// Test that the generated password can be used for authentication
	passwordService := service.NewPasswordService()
	err = passwordService.VerifyPassword(result1.User().u.PasswordHash, result1.Password())
	assert.Nil(s.T(), err) // Should verify successfully
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
