package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"

	"github.com/ystv/stv-web/auth"
	"github.com/ystv/stv-web/controllers"
	utilMail "github.com/ystv/stv-web/mail"
	"github.com/ystv/stv-web/middleware"
	"github.com/ystv/stv-web/structs"
)

//go:embed public/*
var embeddedFiles embed.FS

type (
	Router struct {
		config structs.Config
		repos  *controllers.Repos
		router *echo.Echo
		mailer *utilMail.Mailer
		auth   []echo.MiddlewareFunc
	}
	NewRouter struct {
		Config structs.Config
		Repos  *controllers.Repos
		Debug  bool
		Mailer *utilMail.Mailer
		Auth   []echo.MiddlewareFunc
	}
)

func New(conf NewRouter) *Router {
	r := &Router{
		config: conf.Config,
		router: echo.New(),
		repos:  conf.Repos,
		mailer: conf.Mailer,
		auth:   conf.Auth,
	}
	r.router.HideBanner = true

	r.router.Debug = r.config.Server.Debug

	middleware.New(r.router, r.config.Server.DomainName)

	r.loadRoutes()

	return r
}

func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(":" + r.config.Server.Port))
	return fmt.Errorf("failed to start router on port %s", r.config.Server.Address)
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.repos.Error.Error404)

	r.router.HTTPErrorHandler = r.repos.Error.CustomHTTPErrorHandler

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.Use(r.auth...)

	r.router.GET("/", r.repos.Home.Home)

	authGroup := r.router.Group("/oauth")
	{
		authGroup.GET("", r.repos.Error.Error404)
		authGroup.GET("/callback", auth.NewOauthCallback(r.config.Keycloak))
		authGroup.GET("/logout", auth.NewOauthLogout(r.config.Keycloak))
	}

	admin := r.router.Group("/admin")
	{
		admin.GET("", r.repos.Admin.Admin)
		admin.GET("/elections", r.repos.Admin.Elections)
		election := admin.Group("/election")
		{
			election.GET("/:id", r.repos.Admin.Election)
			election.POST("", r.repos.Admin.AddElection)
			election.POST("/edit/:id", r.repos.Admin.EditElection)
			election.POST("/exclude/:id", r.repos.Admin.Exclude)
			election.POST("/include/:id/:email", r.repos.Admin.Include)
			election.POST("/open/:id", r.repos.Admin.OpenElection)
			election.POST("/close/:id", r.repos.Admin.CloseElection)
			election.POST("/delete/:id", r.repos.Admin.DeleteElection)
			candidates := election.Group("/candidate")
			{
				candidates.POST("/:id", r.repos.Admin.AddCandidate)
				candidates.POST("/delete/:id", r.repos.Admin.DeleteCandidate)
			}
		}
		voters := admin.Group("/voters")
		{
			voters.GET("", r.repos.Admin.Voters)
			// voters.POST("", r.repos.Admin.AddVoter)
			voters.POST("/delete", r.repos.Admin.DeleteVoter)
			voters.POST("/registration", r.repos.Admin.SwitchRegistration)
		}
		admin.POST("/"+r.config.Server.ForceResetURLEndpoint, r.repos.Admin.ForceReset)
	}

	registration := r.router.Group("/registration")
	{
		registration.GET("", r.repos.Registration.Register)
		registration.GET("/qr", r.repos.Registration.QR)
	}

	vote := r.router.Group("/vote/:url")
	{
		vote.GET("", r.repos.Vote.Vote)
		vote.POST("", r.repos.Vote.AddVote)
	}

	r.router.GET("/api/health", func(c echo.Context) error {
		marshal, err := json.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			log.Println(err)
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  err.Error(),
				Internal: err,
			}
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, marshal)
	})

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "public")))

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))
}
