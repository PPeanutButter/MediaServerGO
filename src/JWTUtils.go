package main

import (
	"github.com/cristalhq/jwt/v4"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"log"
	"net/http"
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
	signer, err := jwt.NewSignerHS(jwt.Algorithm(cfg.JWT.Algorithm), []byte(cfg.JWT.Secret))
	if err != nil {
		log.Println("GenerateToken", "NewSignerHS", err)
		return "", err
	}
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(claims)
	if err != nil {
		log.Println("GenerateToken", "Build", err)
		return "", err
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
	verifier, err := jwt.NewVerifierHS(jwt.Algorithm(cfg.JWT.Algorithm), []byte(cfg.JWT.Secret))
	if err != nil {
		log.Println("ParseToken", "NewVerifierHS", err)
		return nil, err
	}
	newToken, err := jwt.Parse([]byte(tokenString), verifier)
	if err != nil {
		log.Println("ParseToken", "Parse", err)
		return nil, err
	}
	err = verifier.Verify(newToken)
	if err != nil {
		log.Println("ParseToken", "Verify", err)
		return nil, err
	}
	var newClaims MyClaims
	err = json.Unmarshal(newToken.Claims(), &newClaims)
	if err != nil {
		log.Println("ParseToken", "Unmarshal", err)
		return nil, err
	}
	return &newClaims, nil
}

func withUser(user string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := getToken(c)
		if len(auth) != 0 {
			body, err := ParseToken(auth, config)
			if err != nil && body.UserName == user {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
}
