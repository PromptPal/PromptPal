package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type userTestSuite struct {
	suite.Suite
	router *gin.Engine
	w3     *service.MockWeb3Service
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *userTestSuite) SetupTest() {
	config.SetupConfig(true)
	w3 := service.NewMockWeb3Service(s.T())
	iai := service.NewMockIsomorphicAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	s.w3 = w3
	s.router = SetupGinRoutes("test", w3, iai, hs, nil)
}

func (s *userTestSuite) GetAuthToken() (result authResponse, err error) {
	w := httptest.NewRecorder()
	payload := `{"address": "0x4-routes-user_test", "signature": "signature", "message": "message"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &result)
	return
}

func (s *userTestSuite) TestAuthMethod() {
	s.w3.On(
		"VerifySignature",
		"0x4-routes-user_test",
		"message",
		"signature",
	).Return(true, nil)

	result, err := s.GetAuthToken()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "0x4-routes-user_test", result.User.Addr)
	assert.NotEmpty(s.T(), result.Token)
}

// Helper method to create a test user with password
func (s *userTestSuite) CreateTestUserWithPassword(username, email, password string) *ent.User {
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
func (s *userTestSuite) CreateTestUserWithoutPassword(username, email string) *ent.User {
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

func (s *userTestSuite) TestPasswordAuthWithUsername() {
	// Create test user with password
	testUser := s.CreateTestUserWithPassword("ahh", "test@example.com", "validpassword123")

	w := httptest.NewRecorder()
	payload := `{"username": "ahh", "password": "validpassword123"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 200, w.Code)

	var result authResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), testUser.ID, result.User.ID)
	assert.NotEmpty(s.T(), result.Token)

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *userTestSuite) TestPasswordAuthWithEmail() {
	// Create test user with password
	testUser := s.CreateTestUserWithPassword("emailuser1", "email@example.com", "validpassword123")

	w := httptest.NewRecorder()
	payload := `{"username": "email@example.com", "password": "validpassword123"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 200, w.Code)

	var result authResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), testUser.ID, result.User.ID)
	assert.NotEmpty(s.T(), result.Token)

	service.EntClient.User.DeleteOneID(testUser.ID).Exec(context.Background())
}

func (s *userTestSuite) TestPasswordAuthInvalidCredentials() {
	// Create test user with password
	u := s.CreateTestUserWithPassword("invaliduser2", "invalid@example.com", "validpassword123")

	w := httptest.NewRecorder()
	payload := `{"username": "invaliduser2", "password": "wrongpassword"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 401, w.Code)

	var result errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid credentials", result.ErrorMessage)

	service.EntClient.User.DeleteOneID(u.ID).Exec(context.Background())
}

func (s *userTestSuite) TestPasswordAuthUserNotFound() {
	w := httptest.NewRecorder()
	payload := `{"username": "nonexistentuser", "password": "anypassword"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 401, w.Code)

	var result errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid credentials", result.ErrorMessage)
}

func (s *userTestSuite) TestPasswordAuthUserWithoutPassword() {
	// Create test user without password
	u := s.CreateTestUserWithoutPassword("nopassuser4", "nopass@example.com")

	w := httptest.NewRecorder()
	payload := `{"username": "nopassuser4", "password": "anypassword"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 401, w.Code)

	var result errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid credentials", result.ErrorMessage)

	service.EntClient.User.DeleteOneID(u.ID).Exec(context.Background())
}

func (s *userTestSuite) TestPasswordAuthInvalidRequestFormat() {
	w := httptest.NewRecorder()
	payload := `{"invalid": "json"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 400, w.Code)

	var result errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid request format", result.ErrorMessage)
}

func (s *userTestSuite) TestPasswordAuthMissingFields() {
	w := httptest.NewRecorder()
	payload := `{"username": "ahh"}` // missing password
	req, _ := http.NewRequest("POST", "/api/v1/auth/password-login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), 400, w.Code)

	var result errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "invalid request format", result.ErrorMessage)
}

func (s *userTestSuite) TearDownSuite() {
	service.Close()
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(userTestSuite))
}
