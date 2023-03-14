package controllers

import (
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/utils"
)

type Repos struct {
	Admin  *AdminRepo
	Home   *HomeRepo
	Public *PublicRepo
	Vote   *VoteRepo
}

func NewRepos(controller Controller, mailer *utils.Mailer, store *store.Store) *Repos {
	return &Repos{
		Admin:  NewAdminRepo(controller, mailer, store),
		Home:   NewHomeRepo(controller),
		Public: NewPublicRepo(controller),
		Vote:   NewVoteRepo(controller, store),
	}
}
