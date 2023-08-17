package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type baseTestSuite struct {
	suite.Suite
}

func (s *baseTestSuite) SetupSuite() {
}

func (s *baseTestSuite) TestGraphQLString() {
	schemaString := String()
	assert.NotEmpty(s.T(), schemaString)
}

func (s *baseTestSuite) TearDownSuite() {
}

func TestBaseTestSuite(t *testing.T) {
	suite.Run(t, new(baseTestSuite))
}
