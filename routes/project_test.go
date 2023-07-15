package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type projectTestSuite struct {
	suite.Suite
	router *gin.Engine
	token  string
}

func (s *projectTestSuite) SetupTest() {

	// just for get a token
	u := new(userTestSuite)
	u.SetT(s.T())
	u.SetS(&s.Suite)
	u.SetupTest()
	authInfo, _ := u.GetAuthToken()
	s.token = authInfo.Token
	u.TearDownSuite()

	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	s.router = SetupGinRoutes("test", w3, oi, hs)
}

func (s *projectTestSuite) TestCreateProject() {
	w := httptest.NewRecorder()
	payload := `{"name": "iuiu", "openaiToken": "openaiToken"}`
	req, _ := http.NewRequest("POST", "/api/v1/admin/projects", strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ent.Project{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "iuiu", result.Name)
	assert.Equal(s.T(), "https://api.openai.com/v1", result.OpenAIBaseURL)
	assert.NotEmpty(s.T(), result.ID)
}

func (s *projectTestSuite) TestListProject() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/projects?limit=10&cursor=999999", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ListResponse[*ent.Project]{}

	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), 1, result.Count)
	assert.Equal(s.T(), 1, len(result.Data))

	d := result.Data[0]

	assert.Equal(s.T(), "iuiu", d.Name)
}

func (s *projectTestSuite) TestGetProject() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/admin/projects/1", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)
	s.router.ServeHTTP(w, req)
	assert.Equal(s.T(), 200, w.Code)

	result := ent.Project{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), "iuiu", result.Name)
}

func (s *projectTestSuite) TearDownSuite() {
	service.Close()
}

func TestProjectTestSuite(t *testing.T) {
	suite.Run(t, new(projectTestSuite))
}
