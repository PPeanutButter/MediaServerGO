package main

import (
	"github.com/cristalhq/jwt/v4"
	"github.com/goccy/go-json"
	"time"
)

type MyClaims struct {
	userName string
	jwt.RegisteredClaims
}

const (
	TokenExpireDuration = time.Hour * 24 * 7
)

func GenerateToken(userName string, cfg Config) (string, error) {
	expirationTime := time.Now().Add(TokenExpireDuration) // 两个小时有效期
	claims := &MyClaims{
		userName: userName,
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
	verifier, err1 := jwt.NewVerifierHS(jwt.Algorithm(cfg.JWT.Algorithm), []byte(cfg.JWT.Secret))
	if err1 != nil {
		return false
	}
	newToken, err2 := jwt.Parse([]byte(tokenString), verifier)
	if err2 != nil {
		return false
	}
	err3 := verifier.Verify(newToken)
	if err3 != nil {
		return false
	}
	var newClaims MyClaims
	errClaims := json.Unmarshal(newToken.Claims(), &newClaims)
	if errClaims != nil {
		return false
	}
	return newClaims.IsValidAt(time.Now())
}
