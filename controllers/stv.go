package controllers

import (
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/utils"
)

type Repos struct {
	Admin        *AdminRepo
	Home         *HomeRepo
	Registration *RegistrationRepo
	Vote         *VoteRepo
}

func NewRepos(controller Controller, mailer *utils.Mailer, store *store.Store, mailConfig utils.MailConfig) *Repos {
	return &Repos{
		Admin:        NewAdminRepo(controller, mailer, store, mailConfig),
		Home:         NewHomeRepo(controller, store),
		Registration: NewRegistrationRepo(controller, store),
		Vote:         NewVoteRepo(controller, store),
	}
}
