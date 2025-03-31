package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type userTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *userTestSuite) SetupTest() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewBaseAIService(s.T())
	gi := mocks.NewBaseAIService(s.T())
	iai := mocks.NewIsomorphicAIService(s.T())
	hs := service.NewHashIDService()

	w3.
		On(
			"VerifySignature",
			"0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
			"message",
			"signature",
		).
		Return(true, nil)

	service.InitDB()
	s.router = SetupGinRoutes("test", w3, oi, gi, iai, hs, nil)
}

func (s *userTestSuite) GetAuthToken() (result authResponse, err error) {
	w := httptest.NewRecorder()
	payload := `{"address": "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8", "signature": "signature", "message": "message"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &result)
	return
}

func (s *userTestSuite) TestAuthMethod() {
	result, err := s.GetAuthToken()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "0x4910c609fbc895434a0a5e3e46b1eb4b64cff2b8", result.User.Addr)
	assert.NotEmpty(s.T(), result.Token)
}

func (s *userTestSuite) TearDownSuite() {
	service.Close()
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(userTestSuite))
}
