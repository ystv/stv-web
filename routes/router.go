package routes

import (
	"embed"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"
	"github.com/ystv/stv_web/auth"
	"github.com/ystv/stv_web/controllers"
	"github.com/ystv/stv_web/middleware"
	"github.com/ystv/stv_web/structs"
	"github.com/ystv/stv_web/utils"
	"io/fs"
	"net/http"
	"net/mail"
)

//go:embed public/*
var embeddedFiles embed.FS

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
	r.router.RouteNotFound("/*", r.repos.Error.Error404)

	r.router.HTTPErrorHandler = r.repos.Error.CustomHTTPErrorHandler

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.GET("/", r.repos.Home.Home)

	admin := r.router.Group("/admin")
	{
		if !r.router.Debug {
			admin.Use(middleware2.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
				if username == r.config.AD.BypassUsername && password == r.config.AD.BypassPassword { // This is here because AD decided to shit the bed
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
					fmt.Printf("error connecting to server: %v\n", err)
					return false, fmt.Errorf("error connecting to server: %w", err)
				}
				defer conn.Conn.Close()

				status, err := conn.Bind(r.config.AD.Bind.Username, r.config.AD.Bind.Password)
				if err != nil {
					fmt.Printf("error binding to server: %v\n", err)
					return false, fmt.Errorf("error binding to server: %w", err)
				}

				if !status {
					return false, fmt.Errorf("error binding to server: invalid credentials")
				}

				status1, err := auth.Authenticate(config, username, password)
				if err != nil {
					fmt.Printf("unable to authenticate %s with error: %v\n", username, err)
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
						fmt.Printf("error getting user groups: %v\n", err)
						return false, fmt.Errorf("error getting user groups: %w", err)
					}

					dnGroups := entry.GetAttributeValues("memberOf")

					if len(dnGroups) == 0 {
						fmt.Println("BIND_SAM user not member of any groups")
						return false, fmt.Errorf("BIND_SAM user not member of any groups")
					}

					stv := false

					for _, group := range dnGroups {
						if group == "CN=STV Admin,CN=Users,DC=ystv,DC=local" {
							stv = true
							return true, nil
						}
					}

					if !stv {
						fmt.Printf("STV not allowed for %s!\n", username)
						return false, fmt.Errorf("STN not allowed for %s!\n", username)
					}
					return true, nil
				} else {
					fmt.Printf("user not authenticated: %s!\n", username)
					return false, fmt.Errorf("user not authenticated: %s!\n", username)
				}
			}))
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

	assetHandler := http.FileServer(getFileSystem())

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))
}

func getFileSystem() http.FileSystem {
	fsys, err := fs.Sub(embeddedFiles, "public")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
