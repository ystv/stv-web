package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv_web/storage"
	"github.com/ystv/stv_web/store"
	"strconv"
)

type VoteRepo struct {
	controller Controller
	store      *store.Store
}

func NewVoteRepo(controller Controller, store *store.Store) *VoteRepo {
	return &VoteRepo{
		controller: controller,
		store:      store,
	}
}

func (r *VoteRepo) Vote(c echo.Context) error {
	url := c.Param("url")
	if len(url) == 0 {
		err := r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Invalid URL"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid url")
	}
	u1, err := r.store.FindURL(url)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Invalid URL"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return err
	}

	e1, err := r.store.FindElection(u1.Election)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Invalid Election"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return err
	}

	if e1.Closed {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Election has been closed"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to vote on a closed election")
	}
	if !e1.Open {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Election has not been opened"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to vote on a non-open election")
	}

	c1, err := r.store.GetCandidatesElectionId(e1.Id)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Candidates cannot be found"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to get candidates")
	}

	v1, err := r.store.FindVoter(u1.Voter)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "Voter cannot be found"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to get voter")
	}

	if u1.Voted {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, nil, "voted.tmpl")
		if err != nil {
			return err
		}
		return nil
	}

	data := struct {
		Election   *storage.Election
		Candidates []*storage.Candidate
		Voter      *storage.Voter
		URL        string
	}{
		Election:   e1,
		Candidates: c1,
		Voter:      v1,
		URL:        u1.Url,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, data, "vote.tmpl")
	if err != nil {
		return err
	}
	return nil
}

func (r *VoteRepo) AddVote(c echo.Context) error {
	url := c.Param("url")

	u1, err := r.store.FindURL(url)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: "URL not found"}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return fmt.Errorf("url not found")
	}

	err = c.Request().ParseForm()
	if err != nil {
		return err
	}

	m := make(map[uint64]string)

	for i := 0; i < len(c.Request().Form); i++ {
		m[uint64(i)] = c.Request().Form.Get("order~" + strconv.FormatUint(uint64(i), 10))
		//fmt.Println("order~"+strconv.FormatUint(uint64(i), 10), c.Request().Form.Get("order~"+strconv.FormatUint(uint64(i), 10)))
	}

	ballot := &storage.Ballot{
		Election: u1.Election,
		Voter:    u1.Voter,
		Choice:   m,
	}

	_, err = r.store.AddBallot(ballot)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, struct{ Error string }{Error: err.Error()}, "voteError.tmpl")
		if err != nil {
			return err
		}
		return err
	}

	err = r.store.SetURLVoted(u1.Url)
	if err != nil {
		return err
	}
	return nil
}
