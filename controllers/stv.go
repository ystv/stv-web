package controllers

type Repos struct {
	Admin  *AdminRepo
	Home   *HomeRepo
	Public *PublicRepo
	Vote   *VoteRepo
}

func NewRepos(controller Controller) *Repos {
	return &Repos{
		Admin:  NewAdminRepo(controller),
		Home:   NewHomeRepo(controller),
		Public: NewPublicRepo(controller),
		Vote:   NewVoteRepo(controller),
	}
}
