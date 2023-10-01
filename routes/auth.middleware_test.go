package routes

import (
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type authMiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
	su     *userTestSuite
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *authMiddlewareTestSuite) SetupTest() {
	config.SetupConfig(true)
	// w3 := mocks.NewWeb3Service(s.T())
	// oi := mocks.NewOpenAIService(s.T())
	// hs := service.NewHashIDService()

	// w3.
	// 	On(
	// 		"VerifySignature",
	// 		"0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
	// 		"message",
	// 		"signature",
	// 	).
	// 	Return(true, nil)

	service.InitDB()

	// su := new(userTestSuite)
	// su.SetT(s.T())
	// su.SetupTest()

	// s.su = su
	// s.router = SetupGinRoutes("test", w3, oi, hs, nil)
}

// func (s *authMiddlewareTestSuite) GetAuthToken() (result authResponse, err error) {
// 	w := httptest.NewRecorder()
// 	payload := `{"address": "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8", "signature": "signature", "message": "message"}`
// 	// req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(payload))
// 	req, _ := http.NewRequest("POST", "/api/v1/admin/prompts/test", strings.NewReader(payload))
// 	req.Header.Add("Content-Type", "application/json")
// 	s.router.ServeHTTP(w, req)
// 	assert.Equal(s.T(), 200, w.Code)
// 	err = json.Unmarshal(w.Body.Bytes(), &result)
// 	return
// }

func (s *authMiddlewareTestSuite) TestAuthMiddleware() {
	// w := httptest.NewRecorder()
	// result, err := s.su.GetAuthToken()

	// assert.Nil(s.T(), err)
	// payload := ``

	// req, _ := http.NewRequest("POST", "/api/v1/admin/prompts/test", strings.NewReader(payload))
	// req.Header.Add("Content-Type", "application/json")
	// s.router.ServeHTTP(w, req)
	// assert.Equal(s.T(), 200, w.Code)
	// err = json.Unmarshal(w.Body.Bytes(), &result)

	// assert.Nil(s.T(), err)
}

func (s *authMiddlewareTestSuite) TestAPIMiddleware() {
	// assert.True(s.T(), false)
	// result, err := s.su.GetAuthToken()
	// assert.Nil(s.T(), err)
	// assert.Equal(s.T(), "0x4910c609fbc895434a0a5e3e46b1eb4b64cff2b8", result.User.Addr)
	// assert.NotEmpty(s.T(), result.Token)
}

func (s *authMiddlewareTestSuite) TearDownSuite() {
	service.Close()
}

func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(authMiddlewareTestSuite))
}
