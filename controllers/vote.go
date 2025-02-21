package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ystv/stv_web/storage"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/templates"
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
		err := r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Invalid URL"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid url")
	}
	u1, err := r.store.FindURL(url)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Invalid URL"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return err
	}

	e1, err := r.store.FindElection(u1.GetElection())
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Invalid Election"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return err
	}

	if e1.GetClosed() {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Election has been closed"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to vote on a closed election")
	}
	if !e1.GetOpen() {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Election has not been opened"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to vote on a non-open election")
	}

	c1, err := r.store.GetCandidatesElectionID(e1.GetId())
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Candidates cannot be found"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to get candidates")
	}

	v1, err := r.store.FindVoter(u1.GetVoter())
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "Voter cannot be found"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to get voter")
	}

	if u1.GetVoted() {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.VotedTemplate)
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
		URL:        u1.GetUrl(),
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.VoteTemplate)
	if err != nil {
		return err
	}
	return nil
}

func (r *VoteRepo) AddVote(c echo.Context) error {
	url := c.Param("url")

	u1, err := r.store.FindURL(url)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: "URL not found"}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return fmt.Errorf("url not found")
	}

	if u1.Voted {
		return fmt.Errorf("this url has expired")
	}

	err = c.Request().ParseForm()
	if err != nil {
		return err
	}

	m := make(map[uint64]string)

	var i uint64
	for i = 0; i < uint64(len(c.Request().Form)); i++ {
		m[i] = c.Request().Form.Get("order~" + strconv.FormatUint(i, 10))
	}

	j, err := json.Marshal(m)
	if err != nil {
		return err
	}

	enc, err := r.controller.encrypt(j)
	if err != nil {
		return err
	}

	ballot := &storage.Ballot{
		Election: u1.GetElection(),
		Choice:   fmt.Sprintf("%0x", enc),
	}

	_, err = r.store.AddBallot(ballot)
	if err != nil {
		err = r.controller.Template.RenderTemplate(c.Response().Writer, struct{ Error string }{Error: err.Error()}, templates.VoteErrorTemplate)
		if err != nil {
			return err
		}
		return err
	}

	err = r.store.SetURLVoted(u1.GetUrl())
	if err != nil {
		return err
	}

	err = r.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.VotedTemplate)
	if err != nil {
		return err
	}
	return nil
}
