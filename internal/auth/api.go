package auth

import (
	"context"
	"github.com/caojs/go-template/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/pkg/errors"
	"net/http"
)

var (
	cfg *config.Config
	providers = make(map[string]goth.Provider)
)

func RouterHandler(r gin.IRouter, config *config.Config) {
	cfg = config

	gothConfig(cfg)

	r.GET("/auth/:provider", oauthRequest)
	r.GET("/auth/:provider/callback", oauthCallback)
}

func gothConfig(config *config.Config) {
	googlePvd := google.New(config.Google.ClientID, config.Google.Secret, config.Google.Callback)
	goth.UseProviders(
		googlePvd,
	)

	providers["google"] = googlePvd
}

func oauthRequest(ctx *gin.Context) {
	name := ctx.Param("provider")
	if _, ok := providers[name]; !ok {
		_ = ctx.AbortWithError(http.StatusBadRequest, errors.New("Provider not found"))
		return
	}

	newRequestCxt := context.WithValue(ctx.Request.Context(), "provider", name)
	ctx.Request = ctx.Request.WithContext(newRequestCxt)

	if gothUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request); err == nil {
		ctx.JSON(http.StatusOK, gothUser)
	} else {
		gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
	}
}

func oauthCallback(ctx *gin.Context) {
	user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

