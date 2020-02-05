package auth

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

var (
	TimeToLive = 60 * 60 * time.Second
	TimeToValid = 60 * time.Second
)

type customClaims struct {
	UserID string `json:"user_id"`
	TTL int64 `json:"ttl"`
	jwt.StandardClaims
}

func createToken(userID string) (string, error) {
	claims := customClaims{
		userID,
		time.Now().UTC().Add(TimeToLive).Unix(),
		jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(TimeToValid).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret"))
}

func createCookie(userID string) (*http.Cookie, error) {
	token, err := createToken(userID)
	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name: "jwt",
		Value: token,
		Path: "/",
		MaxAge: int(TimeToLive.Seconds()),
	}, nil
}

func expireCookie() *http.Cookie {
	return &http.Cookie{
		Name:       "jwt",
		Value:      "",
		Path:       "/",
		MaxAge:     -1,
	}
}
