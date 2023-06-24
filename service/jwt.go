package service

import (
	"errors"
	"time"

	"github.com/PromptPal/PromptPal/config"
	"github.com/PromptPal/PromptPal/ent"
	"github.com/golang-jwt/jwt/v5"
)

type adminTokenClaim struct {
	Uid   int   `json:"uid"`
	Level uint8 `json:"level"`
	// jwt.MapClaims
	jwt.RegisteredClaims
}

func SignJWT(u *ent.User, ttl time.Duration) (string, error) {
	key := config.GetRuntimeConfig().JwtToken
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		adminTokenClaim{
			u.ID,
			u.Level,
			jwt.RegisteredClaims{
				Issuer:    "PromptPal",
				Subject:   u.Addr,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Audience:  []string{"web"},
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		})
	s, err := t.SignedString(key)
	return s, err
}
func ParseJWT(token string) (*adminTokenClaim, error) {
	parsed, err := jwt.ParseWithClaims(token, &adminTokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		return config.GetRuntimeConfig().JwtToken, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := parsed.Claims.(*adminTokenClaim); ok && parsed.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
