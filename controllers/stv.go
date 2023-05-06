package controllers

import (
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/utils"
)

type Repos struct {
	Admin        *AdminRepo
	Home         *HomeRepo
	Public       *PublicRepo
	Registration *RegistrationRepo
	Vote         *VoteRepo
}

func NewRepos(controller Controller, mailer *utils.Mailer, store *store.Store) *Repos {
	return &Repos{
		Admin:        NewAdminRepo(controller, mailer, store),
		Home:         NewHomeRepo(controller, store),
		Public:       NewPublicRepo(controller),
		Registration: NewRegistrationRepo(controller, store),
		Vote:         NewVoteRepo(controller, store),
	}
}
