package schema

import (
	"context"
	"strings"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type authTestSuite struct {
	suite.Suite
}

func (s *authTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := service.NewMockWeb3Service(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs, w3)

	w3.
		On(
			"VerifySignature",
			"test-addr-0x4-schema_auth_test_7777",
			"message",
			"signature",
		).
		Return(true, nil)
}

func (s *authTestSuite) TestAuth() {
	user, err := service.EntClient.User.
		Create().
		SetUsername("0x4-schema_auth_test_7777").
		SetEmail("0x4-schema_auth_test_7777@annatarhe.com").
		SetPasswordHash("hash").
		SetAddr("test-addr-0x4-schema_auth_test_7777").
		SetName("Test User").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())
	assert.Nil(s.T(), err)
	q := QueryResolver{}
	res, err := q.Auth(context.Background(), authInput{
		Auth: authAuthData{
			Address:   "test-addr-0x4-schema_auth_test_7777",
			Message:   "message",
			Signature: "signature",
		},
	})

	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(),
		strings.ToLower("test-addr-0x4-schema_auth_test_7777"),
		res.User().Addr(),
	)
	assert.NotEmpty(s.T(), res.Token())

	service.EntClient.User.DeleteOneID(user.ID).Exec(context.Background())
}

// Helper method to create a test user with password
func (s *authTestSuite) CreateTestUserWithPassword(username, email, password string) *ent.User {
	passwordService := service.NewPasswordService()
	hash, err := passwordService.HashPassword(password)
	assert.Nil(s.T(), err)

	user, err := service.EntClient.User.
		Create().
		SetUsername(username).
		SetEmail(email).
		SetPasswordHash(hash).
		SetAddr("test-addr-" + username).
		SetName("Test User").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())

	assert.Nil(s.T(), err)
	return user
}

// Helper method to create a test user without password
func (s *authTestSuite) CreateTestUserWithoutPassword(username, email string) *ent.User {
	user, err := service.EntClient.User.
		Create().
		SetUsername(username).
		SetEmail(email).
		SetAddr("test-addr-" + username).
		SetName("Test User").
		SetPhone("").
		SetLang("en").
		SetLevel(1).
		Save(context.Background())

	assert.Nil(s.T(), err)
	return user
}

func (s *authTestSuite) TestPasswordAuthWithEmail() {
	// Create test user with password
	testUser := s.CreateTestUserWithPassword("graphqluser", "graphql@example.com", "validpassword123")

	q := QueryResolver{}
	res, err := q.PasswordAuth(context.Background(), passwordAuthInput{
		Auth: passwordAuthData{
			Email: "graphql@example.com",
			Password: "validpassword123",
		},
	})

	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), testUser.ID, res.User().ID())
	assert.NotEmpty(s.T(), res.Token())

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *authTestSuite) TestPasswordAuthInvalidEmail() {
	// Create test user with password
	testUser := s.CreateTestUserWithPassword("graphqlemail", "graphqlemail@example.com", "validpassword123")

	q := QueryResolver{}
	_, err := q.PasswordAuth(context.Background(), passwordAuthInput{
		Auth: passwordAuthData{
			Email: "wrongemail@example.com",
			Password: "validpassword123",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid credentials")

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *authTestSuite) TestPasswordAuthInvalidCredentials() {
	// Create test user with password
	testUser := s.CreateTestUserWithPassword("graphqlinvalid", "graphqlinvalid@example.com", "validpassword123")

	q := QueryResolver{}
	_, err := q.PasswordAuth(context.Background(), passwordAuthInput{
		Auth: passwordAuthData{
			Email: "graphqlinvalid@example.com",
			Password: "wrongpassword",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid credentials")

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *authTestSuite) TestPasswordAuthUserNotFound() {
	q := QueryResolver{}
	_, err := q.PasswordAuth(context.Background(), passwordAuthInput{
		Auth: passwordAuthData{
			Email: "nonexistent@example.com",
			Password: "anypassword",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid credentials")
}

func (s *authTestSuite) TestPasswordAuthUserWithoutPassword() {
	// Create test user without password
	testUser := s.CreateTestUserWithoutPassword("graphqlnopass", "graphqlnopass@example.com")

	q := QueryResolver{}
	_, err := q.PasswordAuth(context.Background(), passwordAuthInput{
		Auth: passwordAuthData{
			Email: "graphqlnopass@example.com",
			Password: "anypassword",
		},
	})

	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid credentials")

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *authTestSuite) TearDownSuite() {
	service.Close()
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}
