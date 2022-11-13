package main

import (
	"github.com/cristalhq/jwt/v4"
	"github.com/goccy/go-json"
	"time"
)

type MyClaims struct {
	UserName string
	jwt.RegisteredClaims
}

func GenerateToken(userName string, cfg Config) (string, error) {
	expirationTime := time.Now().Add(time.Hour * time.Duration(config.JWT.DurationHours))
	claims := &MyClaims{
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// 生成Token，指定签名算法和claims
	signer, err1 := jwt.NewSignerHS(jwt.Algorithm(cfg.JWT.Algorithm), []byte(cfg.JWT.Secret))
	if err1 != nil {
		return "", err1
	}
	builder := jwt.NewBuilder(signer)
	token, err2 := builder.Build(claims)
	if err2 != nil {
		return "", err2
	}
	return token.String(), nil
}

func VerifyToken(tokenString string, cfg Config) bool {
	newClaims, errClaims := ParseToken(tokenString, cfg)
	if errClaims != nil {
		return false
	}
	return newClaims.IsValidAt(time.Now())
}

func ParseToken(tokenString string, cfg Config) (*MyClaims, error) {
	verifier, err1 := jwt.NewVerifierHS(jwt.Algorithm(cfg.JWT.Algorithm), []byte(cfg.JWT.Secret))
	if err1 != nil {
		return nil, err1
	}
	newToken, err2 := jwt.Parse([]byte(tokenString), verifier)
	if err2 != nil {
		return nil, err2
	}
	err3 := verifier.Verify(newToken)
	if err3 != nil {
		return nil, err3
	}
	var newClaims MyClaims
	errClaims := json.Unmarshal(newToken.Claims(), &newClaims)
	if errClaims != nil {
		return nil, errClaims
	}
	return &newClaims, nil
}
