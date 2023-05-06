package routes

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"
	"github.com/ystv/stv_web/auth"
	"github.com/ystv/stv_web/controllers"
	"github.com/ystv/stv_web/middleware"
	"github.com/ystv/stv_web/structs"
	"github.com/ystv/stv_web/utils"
	"net/http"
	"net/mail"
)

type (
	Router struct {
		config *structs.Config
		port   string
		repos  *controllers.Repos
		router *echo.Echo
		mailer *utils.Mailer
	}
	NewRouter struct {
		Config *structs.Config
		Port   string
		Repos  *controllers.Repos
		Debug  bool
		Mailer *utils.Mailer
	}
)

func New(conf *NewRouter) *Router {
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
	r.router.Logger.Error(r.router.Start(r.config.Server.Port))
	return fmt.Errorf("failed to start router on port %s", r.config.Server.Port)
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, utils.Error{Error: "Not found"})
	})

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.GET("/", r.repos.Home.Home)

	admin := r.router.Group("/admin")
	{
		if !r.router.Debug {
			admin.Use(middleware2.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
				config := &auth.Config{
					Server:   r.config.AD.Server,
					Port:     r.config.AD.Port,
					BaseDN:   r.config.AD.BaseDN,
					Security: auth.SecurityType(r.config.AD.Security),
				}

				conn, err := config.Connect()
				if err != nil {
					return false, fmt.Errorf("error connecting to server: %w", err)
				}
				defer conn.Conn.Close()

				status, err := conn.Bind(r.config.AD.Bind.Username, r.config.AD.Bind.Password)
				if err != nil {
					return false, fmt.Errorf("error binding to server: %w", err)
				}

				if !status {
					return false, fmt.Errorf("error binding to server: invalid credentials")
				}

				status1, err := auth.Authenticate(config, username, "Password123")
				if err != nil {
					return false, fmt.Errorf("unable to authenticate %s with error: %w", username, err)
				}

				if status1 {
					var entry *ldap.Entry
					if _, err = mail.ParseAddress(username); err == nil {
						entry, err = conn.GetAttributes("userPrincipalName", username, []string{"memberOf"})
					} else {
						entry, err = conn.GetAttributes("samAccountName", username, []string{"memberOf"})
					}
					if err != nil {
						return false, fmt.Errorf("error getting user groups: %w", err)
					}

					dnGroups := entry.GetAttributeValues("memberOf")

					if len(dnGroups) == 0 {
						return false, fmt.Errorf("BIND_SAM user not member of any groups")
					}

					stv := false

					for _, group := range dnGroups {
						if group == "CN=STV Admin,CN=Users,DC=ystv,DC=local" {
							stv = true
							return false, fmt.Errorf("STV allowed for %s!\n", username)
						}
					}

					if !stv {
						return false, fmt.Errorf("STN not allowed for %s!\n", username)
					}
					return true, nil
				} else {
					return false, fmt.Errorf("user not authenticated: %s!\n", username)
				}
			}))
		}
		admin.GET("", r.repos.Admin.Admin)
		admin.GET("/elections", r.repos.Admin.Elections)
		election := admin.Group("/election")
		{
			election.GET("/:id", r.repos.Admin.Election)
			election.PUT("", r.repos.Admin.AddElection)
			election.PATCH("/:id", r.repos.Admin.EditElection)
			election.PATCH("/open/:id", r.repos.Admin.OpenElection)
			election.PATCH("/close/:id", r.repos.Admin.CloseElection)
			election.DELETE("", r.repos.Admin.DeleteElection)
			candidates := election.Group("/candidate")
			{
				candidates.PUT("/:id", r.repos.Admin.AddCandidate)
				candidates.DELETE("", r.repos.Admin.DeleteCandidate)
			}
		}
		voters := admin.Group("/voters")
		{
			voters.GET("", r.repos.Admin.Voters)
			voters.PUT("", r.repos.Admin.AddVoter)
			voters.DELETE("", r.repos.Admin.DeleteVoter)
		}
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
		vote.PUT("", r.repos.Vote.AddVote)
	}

	r.router.GET("/public/:file", r.repos.Public.Public)

	r.router.GET("/public/webfonts/Arial/:file", r.repos.Public.PublicFontArial)

	r.router.GET("/public/webfonts/Allerta/:file", r.repos.Public.PublicFontAllerta)
}
