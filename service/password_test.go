package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type passwordTestSuite struct {
	suite.Suite
	passwordService *PasswordService
}

func (s *passwordTestSuite) SetupTest() {
	s.passwordService = NewPasswordService()
}

func (s *passwordTestSuite) TestHashPassword() {
	password := "validpassword123"
	
	// Test successful hash
	hash, err := s.passwordService.HashPassword(password)
	assert.Nil(s.T(), err)
	assert.NotEmpty(s.T(), hash)
	assert.NotEqual(s.T(), password, hash)
	
	// Hash should be different each time
	hash2, err := s.passwordService.HashPassword(password)
	assert.Nil(s.T(), err)
	assert.NotEqual(s.T(), hash, hash2)
}

func (s *passwordTestSuite) TestHashPasswordTooShort() {
	password := "12345" // 5 characters, below minimum of 6
	
	hash, err := s.passwordService.HashPassword(password)
	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), hash)
	assert.Contains(s.T(), err.Error(), "password must be at least 6 characters long")
}

func (s *passwordTestSuite) TestHashPasswordTooLong() {
	// Create a password longer than 128 characters
	password := ""
	for i := 0; i < 130; i++ {
		password += "a"
	}
	
	hash, err := s.passwordService.HashPassword(password)
	assert.NotNil(s.T(), err)
	assert.Empty(s.T(), hash)
	assert.Contains(s.T(), err.Error(), "password must be at most 128 characters long")
}

func (s *passwordTestSuite) TestVerifyPassword() {
	password := "validpassword123"
	hash, err := s.passwordService.HashPassword(password)
	assert.Nil(s.T(), err)
	
	// Test successful verification
	err = s.passwordService.VerifyPassword(hash, password)
	assert.Nil(s.T(), err)
	
	// Test failed verification with wrong password
	err = s.passwordService.VerifyPassword(hash, "wrongpassword")
	assert.NotNil(s.T(), err)
	
	// Test with empty hash
	err = s.passwordService.VerifyPassword("", password)
	assert.NotNil(s.T(), err)
	
	// Test with invalid hash
	err = s.passwordService.VerifyPassword("invalid-hash", password)
	assert.NotNil(s.T(), err)
}

func (s *passwordTestSuite) TestValidatePassword() {
	// Test valid password
	err := s.passwordService.ValidatePassword("validpass")
	assert.Nil(s.T(), err)
	
	// Test minimum length boundary (exactly 6 characters)
	err = s.passwordService.ValidatePassword("123456")
	assert.Nil(s.T(), err)
	
	// Test maximum length boundary (exactly 128 characters)
	password128 := ""
	for i := 0; i < 128; i++ {
		password128 += "a"
	}
	err = s.passwordService.ValidatePassword(password128)
	assert.Nil(s.T(), err)
	
	// Test too short (5 characters)
	err = s.passwordService.ValidatePassword("12345")
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "password must be at least 6 characters long")
	
	// Test too long (129 characters)
	password129 := ""
	for i := 0; i < 129; i++ {
		password129 += "a"
	}
	err = s.passwordService.ValidatePassword(password129)
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "password must be at most 128 characters long")
	
	// Test empty password
	err = s.passwordService.ValidatePassword("")
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "password must be at least 6 characters long")
}

func (s *passwordTestSuite) TestGenerateRandomPassword() {
	// Test default minimum length
	password, err := s.passwordService.GenerateRandomPassword(6)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 6, len(password))
	
	// Test custom length
	password, err = s.passwordService.GenerateRandomPassword(12)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 12, len(password))
	
	// Test that passwords are different
	password1, err1 := s.passwordService.GenerateRandomPassword(10)
	password2, err2 := s.passwordService.GenerateRandomPassword(10)
	assert.Nil(s.T(), err1)
	assert.Nil(s.T(), err2)
	assert.NotEqual(s.T(), password1, password2)
	
	// Test below minimum length (should adjust to minimum)
	password, err = s.passwordService.GenerateRandomPassword(3)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 6, len(password)) // Should be adjusted to minimum
	
	// Test above maximum length (should adjust to maximum)
	password, err = s.passwordService.GenerateRandomPassword(200)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 128, len(password)) // Should be adjusted to maximum
}

func TestPasswordTestSuite(t *testing.T) {
	suite.Run(t, new(passwordTestSuite))
}