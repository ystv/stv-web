package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"github.com/go-ldap/ldap/v3"
	auth "github.com/korylprince/go-ad-auth/v3"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"

	"github.com/ystv/stv_web/controllers"
	utilMail "github.com/ystv/stv_web/mail"
	"github.com/ystv/stv_web/middleware"
	"github.com/ystv/stv_web/structs"
)

//go:embed public/*
var embeddedFiles embed.FS

type (
	Router struct {
		config structs.Config
		repos  *controllers.Repos
		router *echo.Echo
		mailer *utilMail.Mailer
	}
	NewRouter struct {
		Config structs.Config
		Repos  *controllers.Repos
		Debug  bool
		Mailer *utilMail.Mailer
	}
)

func New(conf NewRouter) *Router {
	r := &Router{
		config: conf.Config,
		router: echo.New(),
		repos:  conf.Repos,
		mailer: conf.Mailer,
	}
	r.router.HideBanner = true

	r.router.Debug = r.config.Server.Debug

	middleware.New(r.router, r.config.Server.DomainName)

	r.loadRoutes()

	return r
}

func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Server.Address))
	return fmt.Errorf("failed to start router on port %s", r.config.Server.Address)
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.repos.Error.Error404)

	r.router.HTTPErrorHandler = r.repos.Error.CustomHTTPErrorHandler

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.GET("/", r.repos.Home.Home)

	admin := r.router.Group("/admin")
	{
		if !r.router.Debug {
			admin.Use(middleware2.BasicAuth(r.ldapServerAuth))
		}
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
			voters.POST("", r.repos.Admin.AddVoter)
			voters.POST("/delete", r.repos.Admin.DeleteVoter)
			voters.POST("/registration", r.repos.Admin.SwitchRegistration)
		}
		admin.POST("/"+r.config.Server.ForceResetURLEndpoint, r.repos.Admin.ForceReset)
	}

	registration := r.router.Group("/registration")
	{
		registration.GET("", r.repos.Registration.Register)
		registration.GET("/qr", r.repos.Registration.QR)
		registration.POST("", r.repos.Registration.AddVoter)
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

func (r *Router) ldapServerAuth(username, password string, _ echo.Context) (bool, error) {
	if len(r.config.AD.BypassUsername) > 0 &&
		len(r.config.AD.BypassPassword) > 0 &&
		username == r.config.AD.BypassUsername &&
		password == r.config.AD.BypassPassword {
		log.Println("bypass used")
		return true, nil
	}
	config := &auth.Config{
		Server:   r.config.AD.Server,
		Port:     r.config.AD.Port,
		BaseDN:   r.config.AD.BaseDN,
		Security: auth.SecurityType(r.config.AD.Security),
	}

	conn, err := config.Connect()
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error connecting to server: %w", err))
	}
	defer func(Conn *ldap.Conn) {
		err = Conn.Close()
		if err != nil {
			log.Printf("failed to close to LDAP server: %+v", err)
		}
	}(conn.Conn)

	status, err := conn.Bind(r.config.AD.Bind.Username, r.config.AD.Bind.Password)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error binding to server: %w", err))
	}

	if !status {
		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error binding to server: invalid credentials"))
	}

	status1, err := auth.Authenticate(config, username, password)
	if err != nil {
		return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("unable to authenticate %s with error: %w", username, err))
	}

	if status1 {
		var entry *ldap.Entry
		if _, err = mail.ParseAddress(username); err == nil {
			entry, err = conn.GetAttributes("userPrincipalName", username, []string{"memberOf"})
		} else {
			entry, err = conn.GetAttributes("samAccountName", username, []string{"memberOf"})
		}
		if err != nil {
			return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error getting user groups: %w", err))
		}

		dnGroups := entry.GetAttributeValues("memberOf")

		if len(dnGroups) == 0 {
			return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("BIND_SAM user not member of any groups"))
		}

		stvGroup := false

		for _, group := range dnGroups {
			if group == "CN=STV Admin,CN=Users,DC=ystv,DC=local" {
				stvGroup = true
			}
		}

		if !stvGroup {
			return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("STV not allowed for %s", username))
		}
		log.Printf("%s is authenticated", username)
		return true, nil
	}
	return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("user not authenticated: %s", username))
}
