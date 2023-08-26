package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"github.com/ystv/stv_web/controllers"
	"github.com/ystv/stv_web/routes"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/structs"
	_ "github.com/ystv/stv_web/templates"
	"github.com/ystv/stv_web/utils"
	"log"
	"os"
	"strconv"
)

func main() {
	var tomlUsed, local, global bool
	var err error

	config := structs.Config{}

	root := false

	_, err = os.ReadFile("/toml/config.toml")
	if err == nil {
		_, err = toml.DecodeFile("/toml/config.toml", config)
		root = true
	} else {
		_, err = os.ReadFile("./toml/config.toml")
		if err == nil {
			_, err = toml.DecodeFile("./toml/config.toml", config)
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

	if !global && !local && !tomlUsed && stvAddress == "" && domainName == "" {
		log.Fatal("unable to find env files, toml file or env variables")
	} else if tomlUsed {
		log.Println("using toml file")
	} else if !global && !local {
		log.Println("using env variables")
	} else if local && global {
		log.Println("using global and local env files")
	} else if !local {
		log.Println("using global env file")
	} else {
		log.Println("using local env file")
	}

	if !tomlUsed {
		debug, err := strconv.ParseBool(os.Getenv("STV_DEBUG"))
		if err != nil {
			debug = false
		}

		adPort, err := strconv.Atoi(os.Getenv("STV_AD_PORT"))
		if err != nil {
			log.Fatalf("failed to get ad port env: %+v", err)
		}

		adSecurity, err := strconv.Atoi(os.Getenv("STV_AD_SECURITY"))
		if err != nil {
			log.Fatalf("failed to get ad security env: %+v", err)
		}

		mailPort, err := strconv.Atoi(os.Getenv("STV_MAIL_PORT"))
		if err != nil {
			log.Fatalf("failed to get mail port env: %+v", err)
		}

		if !tomlUsed {
			config = structs.Config{
				Server: structs.Server{
					Debug:      debug,
					Address:    os.Getenv("STV_ADDRESS"),
					DomainName: os.Getenv("STV_DOMAIN_NAME"),
				},
				AD: structs.AD{
					BypassUsername: os.Getenv("STV_AD_BYPASS_USERNAME"),
					BypassPassword: os.Getenv("STV_AD_BYPASS_PASSWORD"),
					Server:         os.Getenv("STV_AD_SERVER"),
					Port:           adPort,
					BaseDN:         "",
					Security:       adSecurity,
					Bind: structs.ADBind{
						Username: os.Getenv("STV_AD_BIND_USERNAME"),
						Password: os.Getenv("STV_AD_BIND_PASSWORD"),
					},
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

	var mailer *utils.Mailer
	var mailConfig utils.MailConfig
	if config.Mail.Host != "" {
		mailConfig = utils.MailConfig{
			Host:     config.Mail.Host,
			Port:     config.Mail.Port,
			Username: config.Mail.User,
			Password: config.Mail.Password,
		}

		mailer, err = utils.NewMailer(mailConfig)
		if err != nil {
			log.Printf("failed to connect to mail server: %+v", err)
		} else {
			log.Println("Connected to mail server")

			mailer.KeepAlive = true

			mailer.Defaults = utils.Defaults{
				DefaultTo:   config.Mail.DefaultTo,
				DefaultFrom: "YSTV STV <stv@ystv.co.uk>",
			}
		}
	}

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt)
	//go func() {
	//	for sig := range c {
	//		exitingTemplate := template.New("Exiting Template")
	//		exitingTemplate = template.Must(exitingTemplate.Parse("<body>YSTV STV has been stopped!<br><br>{{if .Debug}}Exit signal: {{.Sig}}<br><br>{{end}}</body>"))
	//
	//		starting := utils.Mail{
	//			Subject:     "YSTV STV has been stopped!",
	//			UseDefaults: true,
	//			Tpl:         exitingTemplate,
	//			TplData: struct {
	//				Debug bool
	//				Sig   os.Signal
	//			}{
	//				Debug: config.Server.Debug,
	//				Sig:   sig,
	//			},
	//		}
	//
	//		err = mailer.SendMail(starting)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		err = mailer.Close()
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		os.Exit(0)
	//	}
	//}()

	if err != nil {
		//if mailer != nil {
		//	err1 := mailer.SendErrorFatalMail(utils.Mail{
		//		Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
		//		UseDefaults: true,
		//	})
		//	if err1 != nil {
		//		fmt.Println(err1)
		//	}
		//}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}

	//startingTemplate := template.New("Startup email")
	//startingTemplate = template.Must(startingTemplate.Parse("<html><body>YSTV STV starting{{if .Debug}} in debug mode!<br><b>Do not run in production! Authentication is disabled!</b>{{else}}!{{end}}<br><br><br><br>If you don't get another email then this has started correctly.</body></html>"))
	//
	//subject := "YSTV STV is starting"

	if config.Server.Debug {
		//subject += " in debug mode"
		log.Println("Debug Mode - Disabled auth - do not run in production!")
	}

	//subject += "!"
	//
	//starting := utils.Mail{
	//	Subject:     subject,
	//	UseDefaults: true,
	//	Tpl:         startingTemplate,
	//	TplData: struct {
	//		Debug bool
	//	}{
	//		Debug: config.Server.Debug,
	//	},
	//}

	newStore, err := store.NewStore(root)
	if err != nil {
		log.Fatal("Failed to create store: ", err)
	}

	_, err = newStore.GetAllowRegistration()
	if err != nil {
		_, err = newStore.SetAllowRegistration(false)
		if err != nil {
			log.Fatal("Failed to initialise allow registration in store: ", err)
		}
	}

	controller := controllers.GetController(config.Server.DomainName)

	router1 := routes.New(routes.NewRouter{
		Config:  config,
		Address: config.Server.Address,
		Repos:   controllers.NewRepos(controller, mailer, newStore, mailConfig),
		Debug:   config.Server.Debug,
		Mailer:  mailer,
	})

	//err = mailer.SendMail(starting)
	//if err != nil {
	//	log.Fatal("Unable to send email")
	//}

	//go noOp(mailer)

	err = router1.Start()
	if err != nil {
		//err1 := mailer.SendErrorFatalMail(utils.Mail{
		//	Error:       fmt.Errorf("the web server couldn't be started: %s... exiting", err),
		//	UseDefaults: true,
		//})
		//if err1 != nil {
		//	fmt.Println(err1)
		//}
		log.Fatalf("The web server couldn't be started!\n\n%s\n\nExiting!", err)
	}
}

//func noOp(mailer *utils.Mailer) {
//	for {
//		err := mailer.SMTPClient.Noop()
//		if err != nil {
//			log.Fatal(err)
//		}
//		time.Sleep(5 * time.Second)
//	}
//}
