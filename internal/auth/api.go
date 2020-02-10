package auth

import (
	"github.com/caojs/go-template/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth/providers/openidConnect"
)

func RouterHandler(r gin.IRouter, config *config.Config, db *sqlx.DB) error {

	basic := NewBasic(db)
	oauth := NewOauth("/", "/login", oauthSave(db))

	openid, err := openidConnect.New(config.Google.ClientID, config.Google.Secret, config.Google.Callback, config.Google.DiscoveryURL)
	if err != nil {
		return err
	}

	err = oauth.use("openid-connect", openid)
	if err != nil {
		return err
	}

	r.POST("/auth/sign-up", gin.WrapF(basic.signUp))
	r.POST("/auth/sign-in", gin.WrapF(basic.signIn))
	r.GET("/auth/logout", gin.WrapF(basic.logout))

	r.GET("/auth/openid-connect", gin.WrapF(oauth.login("openid-connect")))
	r.GET("/auth/openid-connect/callback", gin.WrapF(oauth.callback))

	return nil
}

