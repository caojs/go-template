package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"log"
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

type SaveFunc func(user goth.User) (string, error)

type oauth struct {
	providers map[string]goth.Provider
	failureUrl string
	successUrl string
	save SaveFunc
}

func NewOauth(successUrl, failureUrl string, save SaveFunc) *oauth {
	return &oauth{
		providers: make(map[string]goth.Provider),
		failureUrl: failureUrl,
		successUrl: successUrl,
		save: save,
	}
}

func (o *oauth) use(providerKey string, provider goth.Provider) error {
	if _, ok := o.providers[providerKey]; ok {
		return errors.New(fmt.Sprintf("Provider %s already exists", providerKey))
	}

	o.providers[providerKey] = provider
	goth.UseProviders(provider)
	return nil
}

func (o *oauth) login(providerKey string) http.HandlerFunc {
	if _, ok := o.providers[providerKey]; !ok {
		log.Fatal(fmt.Sprintf("provider key %s not found", providerKey))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), gothic.ProviderParamKey, providerKey))
		if user, err := gothic.CompleteUserAuth(w, r); err == nil {
			o.successHandler(w, r, user)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}
}

func (o *oauth) callback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		// TODO: flash message
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	o.successHandler(w, r, user)
}

func (o *oauth) successHandler(w http.ResponseWriter, r *http.Request, user goth.User) {
	userID, err := o.save(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookie, err := createCookie(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, o.successUrl, http.StatusPermanentRedirect)
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
		Name:"jwt",
		Value:token,
		MaxAge: int(TimeToLive.Seconds()),
	}, nil
}

