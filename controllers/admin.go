package controllers

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv_web/storage"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/utils"
	"github.com/ystv/stv_web/voting"
	"html/template"
	"strconv"
	"unicode"
)

type AdminRepo struct {
	controller Controller
	mailer     *utils.Mailer
	store      *store.Store
}

func NewAdminRepo(controller Controller, mailer *utils.Mailer, store *store.Store) *AdminRepo {
	return &AdminRepo{
		controller: controller,
		mailer:     mailer,
		store:      store,
	}
}

func (r *AdminRepo) Admin(c echo.Context) error {
	elections, err := r.store.GetElections()
	if err != nil {
		return err
	}
	total := len(elections)
	toBeOpened := 0
	open := 0
	closed := 0
	errInt := 0
	for _, e := range elections {
		if !e.Open && !e.Closed {
			toBeOpened++
		} else if e.Open && !e.Closed {
			open++
		} else if !e.Open && e.Closed {
			closed++
		} else {
			errInt++
		}
	}
	temp, err := r.store.GetVoters()
	voters := len(temp)
	data := struct {
		Elections struct {
			ToBeOpened int
			Open       int
			Closed     int
			ErrInt     int
			Total      int
		}
		Voters int
	}{
		Elections: struct {
			ToBeOpened int
			Open       int
			Closed     int
			ErrInt     int
			Total      int
		}{
			ToBeOpened: toBeOpened,
			Open:       open,
			Closed:     closed,
			ErrInt:     errInt,
			Total:      total,
		},
		Voters: voters,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, data, "admin.tmpl")
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) AddCandidate(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	temp := c.Param("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}

	name := c.Request().FormValue("name")

	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return err
	}
	if election.Open {
		return fmt.Errorf("cannot add candidate to open election")
	}
	if election.Closed {
		return fmt.Errorf("cannot add candidate to closed election")
	}

	candidates, err := r.store.GetCandidatesElectionId(id)
	if err != nil {
		return err
	}

	for _, candidate := range candidates {
		if candidate.Name == name {
			return fmt.Errorf("cannot have duplicate candidate")
		}
	}

	candidate := &storage.Candidate{
		Id:       uuid.NewString(),
		Election: id,
		Name:     name,
	}

	_, err = r.store.AddCandidate(candidate)
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) DeleteCandidate(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	id := c.Request().FormValue("id")
	if len(id) == 0 {
		return fmt.Errorf("cannot delete invalid id")
	}
	candidate, err := r.store.FindCandidate(id)
	if err != nil {
		return err
	}
	election, err := r.store.FindElection(candidate.Election)
	if err != nil {
		return err
	}
	if election.Open || election.Closed {
		return fmt.Errorf("cannot delete candidate of open or closed election")
	}
	err = r.store.DeleteCandidate(id)
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) Elections(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return err
	}
	var err1 string
	err = c.Request().ParseForm()
	if err != nil {
		return err
	}
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	elections := stv.Elections
	data := struct {
		Elections []*storage.Election
		Error     string
	}{
		Elections: elections,
		Error:     err1,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, data, "elections.tmpl")
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) Election(c echo.Context) error {
	temp := c.Param("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}
	election, err := r.store.FindElection(id)
	if err != nil {
		return err
	}
	var err1 string
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	candidates, err := r.store.GetCandidatesElectionId(id)
	if err != nil {
		return err
	}
	if len(election.Result) > 0 && election.Result != "R.O.N." {
		candidate, err := r.store.FindCandidate(election.Result)
		if err != nil {
			return err
		}
		election.Result = candidate.Name
	}
	noOfBallots := 0
	if election.Open || election.Closed {
		ballots, err := r.store.GetBallotsElectionId(election.Id)
		if err != nil {
			return err
		}
		noOfBallots = len(ballots)
	}
	voters, err := r.store.GetVoters()
	if err != nil {
		return err
	}
	noOfVoters := len(voters)
	data := struct {
		Election   *storage.Election
		Candidates []*storage.Candidate
		Ballots    int
		Voters     int
		Error      string
	}{
		Election:   election,
		Candidates: candidates,
		Ballots:    noOfBallots,
		Voters:     noOfVoters,
		Error:      err1,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, data, "election.tmpl")
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) AddElection(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	name := c.Request().FormValue("name")
	description := c.Request().FormValue("description")
	tempRon := c.Request().FormValue("ron")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	if len(name) <= 0 || len(description) <= 0 {
		c.Request().Form.Add("error", "Name and Description need to be filled")
		err = r.Elections(c)
		if err != nil {
			return err
		}
	}
	election := &storage.Election{
		Name:        name,
		Description: description,
		Ron:         ron,
	}

	e1, err := r.store.AddElection(election)
	if err != nil && e1 == nil {
		return err
	}

	err = r.Elections(c)
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) EditElection(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}

	temp := c.Param("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}

	name := c.Request().FormValue("name1")
	description := c.Request().FormValue("description")
	tempRon := c.Request().FormValue("ron")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	if len(name) <= 0 || len(description) <= 0 {
		c.Request().Form.Add("error", "Name and Description need to be filled")
		err = r.Elections(c)
		if err != nil {
			return err
		}
	}
	election := &storage.Election{
		Id:          id,
		Name:        name,
		Description: description,
		Ron:         ron,
	}

	e1, err := r.store.EditElection(election)
	if err != nil && e1 == nil {
		return err
	}

	err = r.Elections(c)
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) OpenElection(c echo.Context) error {
	temp := c.Param("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return err
	}
	if election.Open {
		return fmt.Errorf("cannot open election that is already open")
	}
	if election.Closed {
		return fmt.Errorf("cannot reopen election that has been closed")
	}

	candidates, err := r.store.GetCandidatesElectionId(id)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return fmt.Errorf("cannot open election with no candidates")
	}

	err = r.store.OpenElection(id)
	if err != nil {
		return err
	}

	emailTemplate := template.New("email.tmpl")
	emailTemplate = template.Must(emailTemplate.ParseFiles("templates/email.tmpl"))

	voters, err := r.store.GetVoters()

	for _, voter := range voters {
		url := &storage.URL{
			Url:      uuid.NewString(),
			Election: id,
			Voter:    voter.Email,
			Voted:    false,
		}

		_, err = r.store.AddURL(url)
		if err != nil {
			return err
		}

		file := utils.Mail{
			Subject: "YSTV - Vote for (" + election.Name + ")",
			Tpl:     emailTemplate,
			To:      voter.Email,
			From:    "YSTV Elections <stv@ystv.co.uk>",
			TplData: struct {
				Election struct {
					Name        string
					Description string
				}
				Voter struct {
					Name string
				}
				URL string
			}{
				Election: struct {
					Name        string
					Description string
				}{
					Name:        election.Name,
					Description: election.Description,
				},
				Voter: struct {
					Name string
				}{
					Name: voter.Name,
				},
				URL: "https://" + r.controller.DomainName + "/vote/" + url.Url,
			},
		}

		err = r.mailer.SendMail(file)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func (r *AdminRepo) CloseElection(c echo.Context) error {
	temp := c.Param("id")
	temp1 := []rune(temp)
	for _, r2 := range temp1 {
		if !unicode.IsNumber(r2) {
			return fmt.Errorf("id expects a positive number, the provided is not a positive number")
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return err
	}
	if !election.Open {
		return fmt.Errorf("cannot close election that is not open")
	}
	if election.Closed {
		return fmt.Errorf("cannot reclose election that has been closed")
	}

	ballots, err := r.store.GetBallotsElectionId(id)
	if err != nil {
		return err
	}

	ron := &voting.Candidate{Name: "R.O.N."}

	var candidates []*voting.Candidate
	if election.Ron {
		candidates = append(candidates, ron)
	}

	candidatesStore, err := r.store.GetCandidatesElectionId(id)
	if err != nil {
		return err
	}

	for _, c1 := range candidatesStore {
		candidates = append(candidates, &voting.Candidate{Name: c1.Id})
	}

	var ballotsVoting []*voting.Ballot
	for _, ballot := range ballots {
		var c2 []*voting.Candidate
		for i := uint64(0); i < uint64(len(ballot.Choice)); i++ {
			choice := ballot.Choice[i]
			for _, c3 := range candidates {
				if c3.Name == choice {
					c2 = append(c2, c3)
				}
			}
		}
		ballotsVoting = append(ballotsVoting, voting.NewBallot(c2))
	}

	electionResults, err := voting.SingleTransferableVote(candidates, ballotsVoting, 1, voting.DefaultSingleTransferableVoteOptions())
	if err != nil {
		return fmt.Errorf("election failed: %v", err)
	}

	winners := electionResults.GetWinners()

	if len(winners) != 1 {
		return fmt.Errorf("invalid abount of winners")
	}

	election.Result = winners[0].Name

	election, err = r.store.EditElection(election)
	if err != nil {
		return err
	}

	err = r.store.CloseElection(id)
	if err != nil {
		return err
	}

	return nil
}

func (r *AdminRepo) DeleteElection(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	temp := c.Request().FormValue("id")
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return err
	}
	err = r.store.DeleteElection(id)
	if err != nil {
		return err
	}
	err = r.Elections(c)
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) Voters(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return err
	}
	voters := stv.Voters
	err = c.Request().ParseForm()
	if err != nil {
		return err
	}
	var err1 string
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	data := struct {
		Voters []*storage.Voter
		Error  string
	}{
		Voters: voters,
		Error:  err1,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, data, "voters.tmpl")
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) AddVoter(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	email := c.Request().FormValue("email")
	name := c.Request().FormValue("name")
	if len(name) <= 0 || len(email) <= 0 {
		c.Request().Form.Add("error", "Name and Email need to be filled")
		err = r.Voters(c)
		if err != nil {
			return err
		}
	}
	_, err = r.store.AddVoter(&storage.Voter{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *AdminRepo) DeleteVoter(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return err
	}
	email := c.Request().FormValue("email")
	err = r.store.DeleteVoter(email)
	if err != nil {
		return err
	}
	return nil
}
