package controllers

import "github.com/labstack/echo/v4"

type VoteRepo struct {
	controller Controller
}

func NewVoteRepo(controller Controller) *VoteRepo {
	return &VoteRepo{
		controller: controller,
	}
}

func (r *VoteRepo) Vote(c echo.Context) error {
	return nil
}
