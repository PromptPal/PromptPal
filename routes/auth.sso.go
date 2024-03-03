package routes

import (
	"net/http"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ssoProviders string

const (
	SsoProviderGoogle ssoProviders = "google"
)

func ssoProviderCheck(c *gin.Context) {
	provider, ok := c.Params.Get("provider")

	// only google sso is supported
	if !ok || provider != string(SsoProviderGoogle) {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid provider: " + provider,
		})
		return
	}

	// check if provider is set
	if ssoGoogle == nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "SSO provider not set. please contract admin",
		})
	}

	c.Set("ssoProvider", ssoProviders(provider))
	c.Next()
}

type ssoLoginArgs struct {
	State string `query:"state"`
}

func ssoLogin(c *gin.Context) {
	provider := c.MustGet("ssoProvider").(ssoProviders)
	if provider != SsoProviderGoogle {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid provider: " + string(provider),
		})
		return
	}

	var params ssoLoginArgs
	if err := c.Bind(&params); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO: check the state
	sess := sessions.Default(c)
	sess.Set("state", params.State)
	sess.Save()

	// redirect to google login
	redirectUrl := ssoGoogle.AuthCodeURL(params.State)

	c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

type ssoCallbackArgs struct {
	State string `query:"state"`
	Code  string `query:"code"`
}

func ssoCallback(c *gin.Context) {
	provider := c.MustGet("ssoProvider").(ssoProviders)
	if provider != SsoProviderGoogle {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid provider: " + string(provider),
		})
		return
	}

	var params ssoCallbackArgs
	if err := c.Bind(&params); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	session := sessions.Default(c)
	retrievedState := session.Get("state")

	if params.State != "" {
		if retrievedState != c.Query("state") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
				ErrorCode:    http.StatusUnauthorized,
				ErrorMessage: "state does not match",
			})
			return
		}
	}

	tok, err := ssoGoogle.Exchange(c.Request.Context(), c.Query("code"))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "missing id_token in token",
		})
		return
	}

	verifier := oidcProvider.VerifierContext(c.Request.Context(), &oidc.Config{
		ClientID: ssoGoogle.ClientID,
	})

	idToken, err := verifier.Verify(c.Request.Context(), rawIDToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	var googleUserInfo SSOGoogleAuth
	if err := idToken.Claims(&googleUserInfo); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	// TODO: check the email is valid or not

	tx, err := service.EntClient.Tx(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	isExist, err := tx.
		User.
		Query().
		Where(
			user.Email(googleUserInfo.Email),
		).
		Exist(c)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}

	var u *ent.User
	if isExist {
		u, err = tx.User.Query().Where(
			user.Email(googleUserInfo.Email),
		).Only(c)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
				ErrorCode:    http.StatusBadRequest,
				ErrorMessage: err.Error(),
			})
			return
		}
	}

	if !isExist {
		// create user
		u, err = tx.
			User.
			Create().
			SetEmail(googleUserInfo.Email).
			SetName(googleUserInfo.Email).
			SetAvatar(googleUserInfo.Picture).
			SetSource("sso:google").
			SetAddr("").
			SetLang("en").
			SetLevel(0).
			SetPhone("").
			Save(c)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{
				ErrorCode:    http.StatusBadRequest,
				ErrorMessage: err.Error(),
			})
			return
		}
	}

	token, err := service.SignJWT(u, time.Hour*24*30)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}
	if err := tx.Commit(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, config.GetRuntimeConfig().PublicDomain+"/sso/cb?token="+token)
}
