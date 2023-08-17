package schema

import (
	"context"
	"testing"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/service"
	"github.com/PromptPal/PromptPal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type authTestSuite struct {
	suite.Suite
}

func (s *authTestSuite) SetupSuite() {
	config.SetupConfig(true)
	w3 := mocks.NewWeb3Service(s.T())
	oi := mocks.NewOpenAIService(s.T())
	hs := service.NewHashIDService()

	service.InitDB()
	Setup(hs, oi, w3)

	w3.
		On(
			"VerifySignature",
			"0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
			"message",
			"signature",
		).
		Return(true, nil)
}

func (s *authTestSuite) TestAuth() {
	q := QueryResolver{}
	res, err := q.Auth(context.Background(), authInput{
		Auth: authAuthData{
			Address:   "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8",
			Message:   "message",
			Signature: "signature",
		},
	})

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "0x4910c609fBC895434a0A5E3E46B1Eb4b64Cff2B8", res.User().Addr())
	assert.NotEmpty(s.T(), res.Token())
}

func (s *authTestSuite) TearDownSuite() {
	service.Close()
}

func TestauthTestSuite(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}
