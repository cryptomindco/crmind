package utils

import (
	"crauth/pkg/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}

func (w *JWTWrapper) CreateAuthClaimSession(loginUser *models.User) (string, *models.AuthClaims, error) {
	authClaims := models.AuthClaims{
		Id:          loginUser.Id,
		Username:    loginUser.Username,
		Expire:      time.Now().Add(time.Hour * time.Duration(w.ExpirationHours)).Unix(),
		Role:        loginUser.Role,
		Createdt:    loginUser.Createdt,
		LastLogindt: loginUser.LastLogindt,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(w.ExpirationHours)).Unix(),
			Issuer:    w.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString([]byte(w.SecretKey))
	if err != nil {
		return "", nil, err
	}
	return tokenString, &authClaims, nil
}

func (w *JWTWrapper) HanlderCheckLogin(bearer string) (*models.AuthClaims, bool) {
	// Should be a bearer token
	if len(bearer) > 6 && strings.ToUpper(bearer[0:7]) == "BEARER " {
		var tokenStr = bearer[7:]
		var claim models.AuthClaims
		_, err := jwt.ParseWithClaims(tokenStr, &claim, func(token *jwt.Token) (interface{}, error) {
			return []byte(w.SecretKey), nil
		})
		if err != nil || claim.Id <= 0 {
			return nil, false
		}
		return &claim, true
	}
	return nil, false
}
