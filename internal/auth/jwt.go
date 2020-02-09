package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type JWT struct {
	secret string
	standardClaims jwt.StandardClaims
}

type customClaims struct {
	Identify string `json:"identify"`
	jwt.StandardClaims
}

func New(secret string, expiredAt int64) *JWT {
	return &JWT{
		secret,
		jwt.StandardClaims{ExpiresAt:expiredAt},
	}
}

func (j *JWT) Sign(identify string) (string, error) {
	claims := customClaims{
		identify,
		j.standardClaims,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
}

func (j *JWT) Parse(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, customClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return "", errors.Wrap(err, "Parsing jwt error")
	}

	if claims, ok := token.Claims.(customClaims); ok && token.Valid {
		return claims.Identify, nil
	} else {
		return "", errors.New("Token is invalid")
	}
}

func (j *JWT) SignCookie(c *gin.Context, identify string) error {
	tokenString, err := j.Sign(identify)
	if err != nil {
		return err
	}

	c.SetCookie(
		"jwt",
		tokenString,
		int(j.standardClaims.ExpiresAt),
		"/",
		"",
		false,
		true,
	)
	return nil
}

func (j *JWT) ParseCookie(c *gin.Context) (string, error) {
	tokenString, err := c.Cookie("jwt")
	if err != nil {
		return "", err
	}

	return j.Parse(tokenString)
}