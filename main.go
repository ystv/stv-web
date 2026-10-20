package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "golang.org/x/crypto/x509roots/fallback" // CA bundle for FROM Scratch

	"github.com/ystv/stv-web/auth"
	"github.com/ystv/stv-web/controllers"
	"github.com/ystv/stv-web/cookie"
	"github.com/ystv/stv-web/mail"
	"github.com/ystv/stv-web/store"
	"github.com/ystv/stv-web/structs"
)

var (
	Commit  = "unknown"
	Version = "unknown"
)

func main() {
	var tomlUsed, local, global bool
	var err error

	config := structs.Config{}

	root := false

	_, err = os.ReadFile("/toml/config.toml")
	if err == nil {
		_, err = toml.DecodeFile("/toml/config.toml", &config)
		root = true
	} else {
		_, err = os.ReadFile("./toml/config.toml")
		if err == nil {
			_, err = toml.DecodeFile("./toml/config.toml", &config)
		}
	}

	tomlUsed = err == nil

	if !tomlUsed {
		err = godotenv.Load(".env") // Load .env
		global = err == nil

		err = godotenv.Overload(".env.local") // Load .env.local
		local = err == nil
	}

	stvAddress := os.Getenv("STV_ADDRESS")
	domainName := os.Getenv("STV_DOMAIN_NAME")

	switch {
	case !global && !local && !tomlUsed && stvAddress == "" && domainName == "":
		log.Fatal("unable to find env files, toml file or env variables")
	case tomlUsed:
		log.Println("using toml file")
	case !global && !local:
		log.Println("using env variables")
	case local && global:
		log.Println("using global and local env files")
	case !local:
		log.Println("using global env file")
	default:
		log.Println("using local env file")
	}

	if !tomlUsed {
		var debug bool
		debug, err = strconv.ParseBool(os.Getenv("STV_DEBUG"))
		if err != nil {
			debug = false
		}

		var mailPort int
		mailPort, err = strconv.Atoi(os.Getenv("STV_MAIL_PORT"))
		if err != nil {
			log.Fatalf("failed to get mail port env: %+v", err)
		}

		if !tomlUsed {
			config = structs.Config{
				Server: structs.Server{
					Debug:      debug,
					Address:    os.Getenv("STV_ADDRESS"),
					DomainName: os.Getenv("STV_DOMAIN_NAME"),
					Port:       os.Getenv("STV_PORT"),
				},
				Keycloak: structs.Keycloak{
					URL:      os.Getenv("STV_KEYCLOAK_URL"),
					Realm:    os.Getenv("STV_KEYCLOAK_REAL"),
					ClientID: os.Getenv("STV_KEYCLOAK_CLIENTID"),
					Secret:   os.Getenv("STV_KEYCLOAK_SECRET"),
				},
				Mail: structs.Mail{
					Host:      os.Getenv("STV_MAIL_HOST"),
					User:      os.Getenv("STV_MAIL_USERNAME"),
					Password:  os.Getenv("STV_MAIL_PASSWORD"),
					Port:      mailPort,
					DefaultTo: os.Getenv("STV_MAIL_DEFAULT_TO"),
				},
			}
		}
	}

	if config.Server.Debug {
		log.SetFlags(log.Llongfile)
		fmt.Println()
		log.Println("running in debug mode, do not use in production")
		fmt.Println()
	}

	var mailer *mail.Mailer
	var mailConfig mail.Config
	if config.Mail.Host != "" {
		mailConfig = mail.Config{
			Host:     config.Mail.Host,
			Port:     config.Mail.Port,
			Username: config.Mail.User,
			Password: config.Mail.Password,
		}

		mailer, err = mail.NewMailer(mailConfig)
		if err != nil {
			log.Printf("failed to connect to mail server: %+v", err)
		} else {
			log.Println("Connected to mail server")

			mailer.KeepAlive = true

			mailer.Defaults = mail.Defaults{
				DefaultTo:   config.Mail.DefaultTo,
				DefaultFrom: "YSTV STV <stv@ystv.co.uk>",
			}
		}
	}

	if err != nil {
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}

	if config.Server.Debug {
		log.Println("Debug Mode - Disabled auth - do not run in production!")
	}

	newStore, err := store.NewStore(root)
	if err != nil {
		log.Fatal("Failed to create store: ", err)
	}

	_, err = newStore.GetAllowRegistration(nil)
	if err != nil {
		_, err = newStore.SetAllowRegistration(false, nil)
		if err != nil {
			log.Fatal("Failed to initialise allow registration in store: ", err)
		}
	}

	controller := controllers.GetController(config.Server.DomainName)

	config.Keycloak.Address = fmt.Sprintf("%v:%v", strings.Trim(config.Server.Address, "/"), config.Server.Port)
	auth_key := sha256.Sum256([]byte(Version + Commit))
	cookie_store := sessions.NewCookieStore(auth_key[:])

	router1 := New(NewRouter{
		Config: config,
		Repos:  controllers.NewRepos(controller, mailer, newStore, mailConfig, Commit, Version),
		Debug:  config.Server.Debug,
		Mailer: mailer,
		Auth: []echo.MiddlewareFunc{
			cookie.SessionMiddleware(cookie_store),
			auth.AuthMw(newStore, config.Keycloak),
		},
	})

	log.Printf("YSTV STV voting site: %s, commit: %s, version: %s\n", config.Server.Address+":"+config.Server.Port, Commit, Version)

	err = router1.Start()
	if err != nil {
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
