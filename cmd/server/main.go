package main

import (
	"fmt"
	"github.com/caojs/go-template/internal/auth"
	"github.com/caojs/go-template/internal/binding"
	cfg "github.com/caojs/go-template/internal/config"
	"github.com/caojs/go-template/internal/erro"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

var (
	configFile string
	config     *cfg.Config
	rootCmd    = &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func main() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file directory")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	conf, err := cfg.New(configFile)
	if err != nil {
		log.Fatal(err)
	}

	config = conf
}

func run() (err error) {
	db, err := sqlx.Open("postgres", "user=postgres dbname=got")
	if err != nil {
		return err
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		var s = new(struct {
			Str   *string                 `json:"str"`
			Num   int                     `json:"num"`
			Num8  int8                    `json:"num8"`
			Slice []string                `json:"slice"`
			Mh    *multipart.FileHeader   `json:"mh"`
			Mhs   []*multipart.FileHeader `json:"mhs"`
		})

		c.Request.ParseMultipartForm(32 << 20)
		if setted, err := binding.Bind(c.Request, s); err != nil {
			fmt.Printf("Error %s", err)
		} else {
			fmt.Printf("Setted %v \n", setted)
			fmt.Printf("%v \n", s)
		}
	})

	r.Use(erro.Handler)
	_ = auth.RouterHandler(r, config, db)

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	return r.Run(strings.Join([]string{config.Host, config.Port}, ":"))
}
