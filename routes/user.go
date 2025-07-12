package routes

import (
	"net/http"
	"strings"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type authResponse struct {
	User  ent.User `json:"user"`
	Token string   `json:"token"`
}

func authHandler(c *gin.Context) {
	payload := authPayload{}
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	// do web3 check
	verified, err := web3Service.VerifySignature(payload.Address, payload.Message, payload.Signature)

	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	if !verified {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid signature",
		})
		return
	}

	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.Addr(strings.ToLower(payload.Address))).
		Only(c)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	// sign web3 token to client
	token, err := service.SignJWT(u, time.Hour*24*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		User:  *u,
		Token: token,
	})
}

type passwordAuthPayload struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func passwordAuthHandler(c *gin.Context) {
	payload := passwordAuthPayload{}
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid request format",
		})
		return
	}

	passwordService := service.NewPasswordService()

	// Find user by email
	u, err := service.EntClient.User.Query().
		Where(user.Email(payload.Email)).
		Only(c)

	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "invalid credentials",
		})
		return
	}

	// Check if user has a password hash
	if u.PasswordHash == "" {
		c.JSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "invalid credentials",
		})
		return
	}

	// Verify password
	if err := passwordService.VerifyPassword(u.PasswordHash, payload.Password); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse{
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "invalid credentials",
		})
		return
	}

	// Generate JWT token
	token, err := service.SignJWT(u, time.Hour*24*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		User:  *u,
		Token: token,
	})
}
