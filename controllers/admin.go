package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"unicode"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/ystv/stv_web/storage"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/templates"
	"github.com/ystv/stv_web/utils"
	"github.com/ystv/stv_web/voting"
)

type AdminRepo struct {
	controller Controller
	mailer     *utils.Mailer
	store      *store.Store
	mailConfig utils.MailConfig
}

func NewAdminRepo(controller Controller, mailer *utils.Mailer, store *store.Store, mailConfig utils.MailConfig) *AdminRepo {
	return &AdminRepo{
		controller: controller,
		mailer:     mailer,
		store:      store,
		mailConfig: mailConfig,
	}
}

func (r *AdminRepo) Admin(c echo.Context) error {
	elections, err := r.store.GetElections()
	if err != nil {
		return r.errorHandle(c, err)
	}
	total := len(elections)
	toBeOpened := 0
	open := 0
	closed := 0
	errInt := 0
	for _, e := range elections {
		switch {
		case !e.Open && !e.Closed:
			toBeOpened++
		case e.Open && !e.Closed:
			open++
		case !e.Open && e.Closed:
			closed++
		default:
			errInt++
		}
	}
	temp, err := r.store.GetVoters()
	if err != nil {
		return r.errorHandle(c, err)
	}
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
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.AdminTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) AddCandidate(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	name := c.Request().FormValue("name")

	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name cannot be empty"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.Open {
		return r.errorHandle(c, fmt.Errorf("cannot add candidate to open election"))
	}
	if election.Closed {
		return r.errorHandle(c, fmt.Errorf("cannot add candidate to closed election"))
	}

	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, candidate := range candidates {
		if candidate.Name == name {
			return r.errorHandle(c, fmt.Errorf("cannot have duplicate candidate"))
		}
	}

	candidate := &storage.Candidate{
		Id:       uuid.NewString(),
		Election: id,
		Name:     name,
	}

	_, err = r.store.AddCandidate(candidate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.election(c, election.Id)
}

func (r *AdminRepo) DeleteCandidate(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("cannot delete invalid id"))
	}
	candidate, err := r.store.FindCandidate(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	election, err := r.store.FindElection(candidate.Election)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.Open || election.Closed {
		return r.errorHandle(c, fmt.Errorf("cannot delete candidate of open or closed election"))
	}
	err = r.store.DeleteCandidate(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.election(c, election.Id)
}

func (r *AdminRepo) Elections(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string
	err = c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
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
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.ElectionsTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) Election(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}
	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.Result != nil {
		if len(election.Result.Winner) > 0 && election.Result.Winner != "R.O.N." {
			var candidate *storage.Candidate
			candidate, err = r.store.FindCandidate(election.Result.Winner)
			if err != nil {
				fmt.Println(err)
				candidate = &storage.Candidate{Name: election.Result.Winner}
			}
			election.Result.Winner = candidate.Name
		}
	}
	noOfBallots := 0
	if election.Open || election.Closed {
		ballots, err := r.store.GetBallotsElectionID(election.Id)
		if err != nil {
			return r.errorHandle(c, err)
		}
		noOfBallots = len(ballots)
	}
	voters, err := r.store.GetVoters()
	if err != nil {
		return r.errorHandle(c, err)
	}
	data := struct {
		Election   *storage.Election
		Candidates []*storage.Candidate
		Ballots    int
		Error      string
		VotersList []*storage.Voter
	}{
		Election:   election,
		Candidates: candidates,
		Ballots:    noOfBallots,
		Error:      err1,
		VotersList: voters,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.ElectionTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) election(c echo.Context, id uint64) error {
	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.Result != nil {
		if len(election.Result.Winner) > 0 && election.Result.Winner != "R.O.N." {
			var candidate *storage.Candidate
			candidate, err = r.store.FindCandidate(election.Result.Winner)
			if err != nil {
				fmt.Println(err)
				candidate = &storage.Candidate{Name: election.Result.Winner}
			}
			election.Result.Winner = candidate.Name
		}
	}
	noOfBallots := 0
	if election.Open || election.Closed {
		ballots, err := r.store.GetBallotsElectionID(election.Id)
		if err != nil {
			return r.errorHandle(c, err)
		}
		noOfBallots = len(ballots)
	}
	voters, err := r.store.GetVoters()
	if err != nil {
		return r.errorHandle(c, err)
	}
	noOfVoters := len(voters)
	data := struct {
		Election   *storage.Election
		Candidates []*storage.Candidate
		Ballots    int
		Voters     int
		Error      string
		URL        string
		VotersList []*storage.Voter
	}{
		Election:   election,
		Candidates: candidates,
		Ballots:    noOfBallots,
		Voters:     noOfVoters - len(election.Excluded),
		Error:      err1,
		URL:        "https://" + r.controller.DomainName + "/admin/election/" + strconv.FormatUint(election.Id, 10),
		VotersList: voters,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.ElectionTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) AddElection(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	name := c.Request().FormValue("name")
	description := c.Request().FormValue("description")
	tempRon := c.Request().FormValue("ron")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and description need to be filled"))
	}
	election := &storage.Election{
		Name:        name,
		Description: description,
		Ron:         ron,
	}

	e1, err := r.store.AddElection(election)
	if err != nil && e1 == nil {
		return r.errorHandle(c, err)
	}
	return r.Elections(c)
}

func (r *AdminRepo) EditElection(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	name := c.Request().FormValue("name1")
	description := c.Request().FormValue("description")
	tempRon := c.Request().FormValue("ron")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and description need to be filled"))
	}
	election := &storage.Election{
		Id:          id,
		Name:        name,
		Description: description,
		Ron:         ron,
	}

	e1, err := r.store.EditElection(election)
	if err != nil && e1 == nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("admin/election/%d", election.Id))
}

func (r *AdminRepo) OpenElection(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.Open {
		return r.errorHandle(c, fmt.Errorf("cannot open election that is already open"))
	}
	if election.Closed {
		return r.errorHandle(c, fmt.Errorf("cannot reopen election that has been closed"))
	}

	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	if len(candidates) == 0 {
		return r.errorHandle(c, fmt.Errorf("cannot open election with no candidates"))
	}

	err = r.store.OpenElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	voters, err := r.store.GetVoters()
	if err != nil {
		return r.errorHandle(c, fmt.Errorf("failed to get voters: %w", err))
	}

	r.mailer, err = utils.NewMailer(r.mailConfig)
	if err != nil {
		log.Printf("failed to reconnect to mail server: %+v", err)
	} else {
		log.Println("Reconnected to mail server")
	}

	election.Voters = uint64(len(voters) - len(election.Excluded))

	go r.sendEmailThread(voters, election)

	return c.Redirect(http.StatusFound, fmt.Sprintf("admin/election/%d", election.Id))
}

func (r *AdminRepo) sendEmailThread(voters []*storage.Voter, election *storage.Election) {
	for _, voter := range voters {
		skip := false
		for _, v := range election.Excluded {
			if voter.Email == v.Email {
				skip = true
			}
		}

		if !skip {
			url := &storage.URL{
				Url:      uuid.NewString(),
				Election: election.Id,
				Voter:    voter.Email,
				Voted:    false,
			}

			_, err := r.store.AddURL(url)
			if err != nil {
				fmt.Println(err)
			}

			file := utils.Mail{
				Subject: "YSTV - Vote for (" + election.Name + ")",
				Tpl:     r.controller.Template.RenderEmail(templates.EmailTemplate),
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
	}
}

func (r *AdminRepo) CloseElection(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if !election.Open {
		return r.errorHandle(c, fmt.Errorf("cannot close election that is not open"))
	}
	if election.Closed {
		return r.errorHandle(c, fmt.Errorf("cannot reclose election that has been closed"))
	}

	ballots, err := r.store.GetBallotsElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	ron := &voting.Candidate{Name: "R.O.N."}

	candidates := make([]*voting.Candidate, 0)
	if election.Ron {
		candidates = append(candidates, ron)
	}

	candidatesStore, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, c1 := range candidatesStore {
		candidates = append(candidates, &voting.Candidate{Name: c1.Id})
	}

	ballotsVoting := make([]*voting.Ballot, 0, len(ballots))
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
		return r.errorHandle(c, fmt.Errorf("election failed: %w", err))
	}

	result := &storage.Result{}

	result.Rounds = uint64(len(electionResults.Rounds))

	for i, round := range electionResults.Rounds {
		rounds := &storage.Round{}
		rounds.Round = uint64(i)
		rounds.Blanks = uint64(round.NumberOfBlankVotes)
		for j, c := range round.CandidateResults {
			candidateStatus := &storage.CandidateStatus{}
			candidateStatus.CandidateRank = uint64(j)
			candidateStatus.Id = c.Candidate.Name
			candidateStatus.NoOfVotes = c.NumberOfVotes
			candidateStatus.Status = string(c.Status)
			rounds.CandidateStatus = append(rounds.CandidateStatus, candidateStatus)
		}
		result.Round = append(result.Round, rounds)
	}
	winners := electionResults.GetWinners()

	if len(winners) != 1 {
		return r.errorHandle(c, fmt.Errorf("invalid abount of winners"))
	}

	result.Winner = winners[0].Name

	election.Result = result

	election, err = r.store.EditElection(election)
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = r.store.CloseElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("admin/election/%d", election.Id))
}

func (r *AdminRepo) Exclude(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}

	email := c.Request().FormValue("email")

	voter, err := r.store.FindVoter(email)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, v := range election.Excluded {
		if v.Email == voter.Email {
			return r.election(c, election.Id)
		}
	}

	election.Excluded = append(election.Excluded, voter)

	_, err = r.store.EditElection(election)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return r.election(c, election.Id)
}

func (r *AdminRepo) Include(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}

	email := c.Param("email")

	voter, err := r.store.FindVoter(email)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for index, v := range election.Excluded {
		if v.Email == voter.Email {
			copy(election.Excluded[index:], election.Excluded[index+1:])     // Shift a[i+1:] left one index
			election.Excluded[len(election.Excluded)-1] = nil                // Erase last element (write zero value)
			election.Excluded = election.Excluded[:len(election.Excluded)-1] // Truncate slice

			_, err = r.store.EditElection(election)
			if err != nil {
				return r.errorHandle(c, err)
			}
			break
		}
	}

	return r.election(c, election.Id)
}

func (r *AdminRepo) DeleteElection(c echo.Context) error {
	temp := c.Param("id")
	for _, r2 := range temp {
		if !unicode.IsNumber(r2) {
			return r.errorHandle(c, fmt.Errorf("id expects a positive number, the provided is not a positive number"))
		}
	}
	id, err := strconv.ParseUint(temp, 10, 64)
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = r.store.DeleteElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return r.Elections(c)
}

func (r *AdminRepo) Voters(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}
	voters := stv.Voters
	err = c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string
	if len(c.Request().FormValue("error")) > 0 {
		err1 = c.Request().FormValue("error")
	}
	data := struct {
		Voters            []*storage.Voter
		AllowRegistration bool
		Error             string
	}{
		Voters:            voters,
		AllowRegistration: stv.AllowRegistration,
		Error:             err1,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.VotersTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) AddVoter(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	email := c.Request().FormValue("email")
	name := c.Request().FormValue("name")
	if len(name) == 0 || len(email) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and email need to be filled"))
	}
	_, err = r.store.AddVoter(&storage.Voter{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.Voters(c)
}

func (r *AdminRepo) DeleteVoter(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	email := c.Request().FormValue("email")
	err = r.store.DeleteVoter(email)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.Voters(c)
}

func (r *AdminRepo) SwitchRegistration(c echo.Context) error {
	allow, err := r.store.GetAllowRegistration()
	if err != nil {
		return r.errorHandle(c, err)
	}
	_, err = r.store.SetAllowRegistration(!allow)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.Voters(c)
}

func (r *AdminRepo) errorHandle(c echo.Context, err error) error {
	data := struct {
		Error string
	}{
		Error: err.Error(),
	}
	fmt.Println(data.Error)
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.AdminErrorTemplate)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
