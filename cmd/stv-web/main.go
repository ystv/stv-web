package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ystv/stv_web/controllers"
	"github.com/ystv/stv_web/routes"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/structs"
	_ "github.com/ystv/stv_web/templates"
	"github.com/ystv/stv_web/utils"
	"html/template"
	"log"
	"os"
	"os/signal"
)

func main() {
	var err error

	config := &structs.Config{}

	_, err = os.ReadFile("/toml/config.toml")
	if err == nil {
		_, err = toml.DecodeFile("/toml/config.toml", config)
	}

	_, err = os.ReadFile("./toml/config.toml")
	if err == nil {
		_, err = toml.DecodeFile("./toml/config.toml", config)
	}

	if err != nil {
		log.Fatal(err)
	}

	if config.Server.Debug {
		log.SetFlags(log.Llongfile)
	}

	var mailer *utils.Mailer
	if config.Mail.Host != "" {
		mailConfig := utils.MailConfig{
			Host:     config.Mail.Host,
			Port:     config.Mail.Port,
			Username: config.Mail.User,
			Password: config.Mail.Password,
		}

		mailer, err = utils.NewMailer(mailConfig)
		if err != nil {
			log.Printf("failed to connect to mail server: %+v", err)
		} else {
			log.Printf("Connected to mail server: %s\n", config.Mail.Host)

			mailer.KeepAlive = true

			mailer.Defaults = utils.Defaults{
				DefaultTo:   "liam.burnand@ystv.co.uk",
				DefaultFrom: "YSTV STV <stv@ystv.co.uk>",
			}
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			exitingTemplate := template.New("Exiting Template")
			exitingTemplate = template.Must(exitingTemplate.Parse("<body>YSTV STV has been stopped!<br><br>{{if .Debug}}Exit signal: {{.Sig}}<br><br>{{end}}</body>"))

			starting := utils.Mail{
				Subject:     "YSTV STV has been stopped!",
				UseDefaults: true,
				Tpl:         exitingTemplate,
				TplData: struct {
					Debug bool
					Sig   os.Signal
				}{
					Debug: config.Server.Debug,
					Sig:   sig,
				},
			}

			err = mailer.SendMail(starting)
			if err != nil {
				fmt.Println(err)
			}
			err = mailer.Close()
			if err != nil {
				fmt.Println(err)
			}
			os.Exit(0)
		}
	}()

	if err != nil {
		if mailer != nil {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
	if err != nil {
		if mailer != nil {
			err1 := mailer.SendErrorFatalMail(utils.Mail{
				Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
				UseDefaults: true,
			})
			if err1 != nil {
				fmt.Println(err1)
			}
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}

	startingTemplate := template.New("Startup email")
	startingTemplate = template.Must(startingTemplate.Parse("<html><body>YSTV STV starting{{if .Debug}} in debug mode!<br><b>Do not run in production! Authentication is disabled!</b>{{else}}!{{end}}<br><br><br><br>If you don't get another email then this has started correctly.</body></html>"))

	subject := "YSTV STV is starting"

	if config.Server.Debug {
		subject += " in debug mode"
		log.Println("Debug Mode - Disabled auth - do not run in production!")
	}

	subject += "!"

	starting := utils.Mail{
		Subject:     subject,
		UseDefaults: true,
		Tpl:         startingTemplate,
		TplData: struct {
			Debug bool
		}{
			Debug: config.Server.Debug,
		},
	}

	newStore, err := store.NewStore()
	if err != nil {
		log.Fatal("Failed to create store", err)
	}

	_, err = newStore.GetAllowRegistration()
	if err != nil {
		_, err = newStore.SetAllowRegistration(false)
		if err != nil {
			log.Fatal("Failed to initialise allow registration in store", err)
		}
	}

	controller := controllers.GetController(config.Server.DomainName)

	router1 := routes.New(&routes.NewRouter{
		Config: config,
		Port:   config.Server.Port,
		Repos:  controllers.NewRepos(controller, mailer, newStore),
		Debug:  config.Server.Debug,
		Mailer: mailer,
	})

	err = mailer.SendMail(starting)
	if err != nil {
		log.Fatal("Unable to send email")
	}

	err = router1.Start()
	if err != nil {
		err1 := mailer.SendErrorFatalMail(utils.Mail{
			Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
			UseDefaults: true,
		})
		if err1 != nil {
			fmt.Println(err1)
		}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}
