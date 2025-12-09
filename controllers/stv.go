package controllers

import (
	"github.com/ystv/stv-web/mail"
	"github.com/ystv/stv-web/store"
)

type Repos struct {
	Admin        *AdminRepo
	Error        *ErrorRepo
	Home         *HomeRepo
	Registration *RegistrationRepo
	Vote         *VoteRepo
}

func NewRepos(controller Controller, mailer *mail.Mailer, store *store.Store, mailConfig mail.Config, commit, version string) *Repos {
	return &Repos{
		Admin:        NewAdminRepo(controller, mailer, store, mailConfig, commit, version),
		Error:        NewErrorRepo(controller),
		Home:         NewHomeRepo(controller, store),
		Registration: NewRegistrationRepo(controller, store),
		Vote:         NewVoteRepo(controller, store),
	}
}
