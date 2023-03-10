package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ystv/stv_web/controllers"
	"github.com/ystv/stv_web/routes"
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
	_, err = toml.DecodeFile("config.toml", config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Server.Debug {
		log.SetFlags(log.Llongfile)
	}

	//access := utils.NewAccesser(utils.Config{
	//	AccessCookieName: config.Server.AccessCookieName, // jwt_token --> base64
	//	DomainName:       config.Server.DomainName,
	//})

	var mailer *utils.Mailer
	if config.Mail.Host != "" {
		if config.Mail.Enabled {
			mailConfig := utils.MailConfig{
				Host:     config.Mail.Host,
				Port:     config.Mail.Port,
				Username: config.Mail.User,
				Password: config.Mail.Password,
			}

			mailer, err = utils.NewMailer(mailConfig)
			if err != nil {
				log.Printf("failed to connect to mail server: %+v", err)
				config.Mail.Enabled = false
			} else {
				log.Printf("Connected to mail server: %s\n", config.Mail.Host)

				mailer.Defaults = utils.Defaults{
					DefaultTo:   "root@bswdi.co.uk",
					DefaultFrom: "YSTV STV <afc@bswdi.co.uk>",
				}
			}
		}
	} else {
		config.Mail.Enabled = false
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if config.Mail.Enabled {
				exitingTemplate := template.New("Exiting Template")
				exitingTemplate = template.Must(exitingTemplate.Parse("<body>YSTV STV has been stopped!<br><br>{{if .Debug}}Exit signal: {{.Sig}}<br><br>{{end}}Version: {{.Version}}<br>Commit: {{.Commit}}</body>"))

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
			}
			os.Exit(0)
		}
	}()

	//err = routes.New(config, access, mailer).Start()
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

	//session, err := handler.NewSession(config.Server.YSTV_API)
	//if err != nil {
	//	if config.Mail.Enabled {
	//		err1 := mailer.SendErrorFatalMail(utils.Mail{
	//			Error:       fmt.Errorf("the session couldn't be initialised: %s... exiting", err),
	//			UseDefaults: true,
	//		})
	//		if err1 != nil {
	//			fmt.Println(err1)
	//		}
	//	}
	//	log.Fatalf("The session couldn't be initialised!\n\n%s\n\nExiting!", err)
	//}

	if config.Mail.Enabled {

		startingTemplate := template.New("Starting Template")
		startingTemplate = template.Must(startingTemplate.Parse("<body>YSTV STV starting{{if .Debug}} in debug mode!<br><b>Do not run in production! Authentication is disabled!</b>{{else}}!{{end}}<br><br>Version: {{.Version}}<br>Commit: {{.Commit}}<br><br>If you don't get another email then this has started correctly.</body>"))

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

		err = mailer.SendMail(starting)
		if err != nil {
			fmt.Println(err)
		}
	}

	controller := controllers.GetController()

	router1 := routes.New(&routes.NewRouter{
		Config: config,
		Port:   config.Server.Port,
		Repos:  controllers.NewRepos(controller),
		//DomainName:
		Debug: config.Server.Debug,
		//Access: accesser,
		Mailer: mailer,
	})
	err = router1.Start()
	if err != nil {
		if config.Mail.Enabled {
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
}
