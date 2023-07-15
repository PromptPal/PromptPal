package routes

// Basic imports
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

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type userTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *userTestSuite) SetupTest() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
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
	s.router = SetupGinRoutes("test", w3, oi, hs)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (s *userTestSuite) TestAuthMethod() {
	w := httptest.NewRecorder()
	payload := `{"address": "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8", "signature": "signature", "message": "message"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := authResponse{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
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
