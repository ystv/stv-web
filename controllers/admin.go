package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	commit     string
	version    string
}

func NewAdminRepo(controller Controller, mailer *utils.Mailer, store *store.Store, mailConfig utils.MailConfig, commit, version string) *AdminRepo {
	return &AdminRepo{
		controller: controller,
		mailer:     mailer,
		store:      store,
		mailConfig: mailConfig,
		commit:     commit,
		version:    version,
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
		case !e.GetOpen() && !e.GetClosed():
			toBeOpened++
		case e.GetOpen() && !e.GetClosed():
			open++
		case !e.GetOpen() && e.GetClosed():
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
		Voters  int
		Commit  string
		Version string
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
		Voters:  voters,
		Commit:  r.commit,
		Version: r.version,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.AdminTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) AddCandidate(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	name := c.FormValue("name")

	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name cannot be empty"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.GetOpen() {
		return r.errorHandle(c, fmt.Errorf("cannot add candidate to open election"))
	}
	if election.GetClosed() {
		return r.errorHandle(c, fmt.Errorf("cannot add candidate to closed election"))
	}

	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, candidate := range candidates {
		if candidate.GetName() == name {
			return r.errorHandle(c, fmt.Errorf("cannot have duplicate candidate"))
		}
	}

	candidate := &storage.Candidate{
		Election: id,
		Name:     name,
	}

	_, err = r.store.AddCandidate(candidate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) DeleteCandidate(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid candidate id"))
	}
	candidate, err := r.store.FindCandidate(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	election, err := r.store.FindElection(candidate.GetElection())
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.GetOpen() || election.GetClosed() {
		return r.errorHandle(c, fmt.Errorf("cannot delete candidate of open or closed election"))
	}
	err = r.store.DeleteCandidate(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) Elections(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string

	if len(c.FormValue("error")) > 0 {
		err1 = c.FormValue("error")
	}
	elections := stv.GetElections()
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
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}
	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	var err1 string
	if len(c.FormValue("error")) > 0 {
		err1 = c.FormValue("error")
	}
	candidates, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.GetResult() != nil {
		if len(election.GetResult().GetWinners()) > 0 {
			var winningCandidates []string
			for _, wc := range election.GetResult().GetWinners() {
				if wc == "R.O.N." {
					winningCandidates = append(winningCandidates, "R.O.N.")
					continue
				}
				var candidate *storage.Candidate
				candidate, err = r.store.FindCandidate(wc)
				if err != nil {
					fmt.Println(err)
					candidate = &storage.Candidate{Name: wc}
				}
				winningCandidates = append(winningCandidates, candidate.GetName())
			}
			election.GetResult().Winners = winningCandidates
		}
	}
	var noOfBallots uint64
	if election.GetOpen() || election.GetClosed() {
		var ballots []*storage.Ballot
		ballots, err = r.store.GetBallotsElectionID(election.GetId())
		if err != nil {
			return r.errorHandle(c, err)
		}
		noOfBallots = uint64(len(ballots))
	}
	voters, err := r.store.GetVoters()
	if err != nil {
		return r.errorHandle(c, err)
	}
	data := struct {
		Election   *storage.Election
		Candidates []*storage.Candidate
		Ballots    uint64
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

func (r *AdminRepo) AddElection(c echo.Context) error {
	name := c.FormValue("name")
	description := c.FormValue("description")
	tempRon := c.FormValue("ron")
	tempSeats := c.FormValue("seats")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	var seats uint64
	seats, err := strconv.ParseUint(tempSeats, 10, 64)
	if err != nil {
		return r.errorHandle(c, fmt.Errorf("number of seats must be an positive integer value between 1 and 3"))
	}
	if seats < 1 || seats > 3 {
		return r.errorHandle(c, fmt.Errorf("number of seats must be an positive integer value between 1 and 3"))
	}
	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and description need to be filled"))
	}
	election := &storage.Election{
		Name:        name,
		Description: description,
		Ron:         ron,
		Seats:       seats,
	}

	e1, err := r.store.AddElection(election)
	if err != nil && e1 == nil {
		return r.errorHandle(c, err)
	}
	return r.Elections(c)
}

func (r *AdminRepo) EditElection(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	name := c.FormValue("name1")
	description := c.FormValue("description")
	tempRon := c.FormValue("ron")
	tempSeats := c.FormValue("seats")
	ron := false
	if len(tempRon) > 0 {
		ron = true
	}
	var seats uint64
	seats, err := strconv.ParseUint(tempSeats, 10, 64)
	if err != nil {
		return r.errorHandle(c, fmt.Errorf("number of seats must be an positive integer value between 1 and 3"))
	}
	if seats < 1 || seats > 3 {
		return r.errorHandle(c, fmt.Errorf("number of seats must be an positive integer value between 1 and 3"))
	}
	if len(name) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and description need to be filled"))
	}
	election := &storage.Election{
		Id:          id,
		Name:        name,
		Description: description,
		Ron:         ron,
		Seats:       seats,
	}

	e1, err := r.store.EditElection(election)
	if err != nil && e1 == nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) OpenElection(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if election.GetOpen() {
		return r.errorHandle(c, fmt.Errorf("cannot open election that is already open"))
	}
	if election.GetClosed() {
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

	//nolint:gosec
	election.Voters = uint64(len(voters) - len(election.GetExcluded()))

	go r.sendEmailThread(voters, election)

	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) sendEmailThread(voters []*storage.Voter, election *storage.Election) {
	for _, voter := range voters {
		fmt.Println(1)
		skip := false
		for _, v := range election.GetExcluded() {
			if voter.GetEmail() == v.GetEmail() {
				skip = true
			}
		}
		fmt.Println(2)

		if !skip {
			fmt.Println(3)
			url := &storage.URL{
				Election: election.GetId(),
				Voter:    voter.GetEmail(),
				Voted:    false,
			}

			url, err := r.store.AddURL(url)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(2, url.GetUrl())

			file := utils.Mail{
				Subject: "YSTV - Vote for (" + election.GetName() + ")",
				Tpl:     r.controller.Template.RenderEmail(templates.EmailTemplate),
				To:      voter.GetEmail(),
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
						Name:        election.GetName(),
						Description: election.GetDescription(),
					},
					Voter: struct {
						Name string
					}{
						Name: voter.GetName(),
					},
					URL: "https://" + r.controller.DomainName + "/vote/" + url.GetUrl(),
				},
			}
			fmt.Println(3, url.GetVoter())

			err = r.mailer.SendMail(file)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (r *AdminRepo) CloseElection(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}
	if !election.GetOpen() {
		return r.errorHandle(c, fmt.Errorf("cannot close election that is not open"))
	}
	if election.GetClosed() {
		return r.errorHandle(c, fmt.Errorf("cannot reclose election that has been closed"))
	}

	if election.Seats < 1 {
		return r.errorHandle(c, fmt.Errorf("cannot close election that has no or negative seats: %d", election.Seats))
	}

	ballots, err := r.store.GetBallotsElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	ron := &voting.Candidate{Name: "R.O.N."}

	candidates := make([]*voting.Candidate, 0)
	if election.GetRon() {
		candidates = append(candidates, ron)
	}

	candidatesStore, err := r.store.GetCandidatesElectionID(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, c1 := range candidatesStore {
		candidates = append(candidates, &voting.Candidate{Name: c1.GetId()})
	}

	ballotsVoting := make([]*voting.Ballot, 0, len(ballots))
	for _, ballot := range ballots {
		var c2 []*voting.Candidate
		var dec []byte
		dec, err = r.controller.decrypt([]byte(ballot.GetChoice()))
		if err != nil {
			return r.errorHandle(c, err)
		}
		m := make(map[uint64]string)
		err = json.Unmarshal(dec, &m)
		if err != nil {
			return r.errorHandle(c, err)
		}
		for _, v := range m {
			fmt.Println(1)
			fmt.Println(v)
			for _, c3 := range candidates {
				fmt.Println(2)
				fmt.Println(c3)
				if c3.Name == v {
					c2 = append(c2, c3)
				}
			}
		}
		ballotsVoting = append(ballotsVoting, voting.NewBallot(c2))
	}

	electionResults, err := voting.SingleTransferableVote(candidates, ballotsVoting, election.Seats, voting.DefaultSingleTransferableVoteOptions())
	if err != nil {
		return r.errorHandle(c, fmt.Errorf("election failed: %w", err))
	}

	result := &storage.Result{}

	result.Rounds = uint64(len(electionResults.Rounds))

	for i, round := range electionResults.Rounds {
		rounds := &storage.Round{}
		//nolint:gosec
		rounds.Round = uint64(i)
		//nolint:gosec
		rounds.Blanks = uint64(round.NumberOfBlankVotes)
		for j, c := range round.CandidateResults {
			candidateStatus := &storage.CandidateStatus{}
			//nolint:gosec
			candidateStatus.CandidateRank = uint64(j)
			candidateStatus.Id = c.Candidate.Name
			candidateStatus.NoOfVotes = c.NumberOfVotes
			candidateStatus.Status = string(c.Status)
			rounds.CandidateStatus = append(rounds.GetCandidateStatus(), candidateStatus)
		}
		result.Round = append(result.GetRound(), rounds)
	}
	winners := electionResults.GetWinners()

	//nolint:gosec
	if uint64(len(winners)) != election.Seats {
		return r.errorHandle(c, fmt.Errorf("invalid abount of winners"))
	}

	names := make([]string, len(winners))
	for i, w := range winners {
		names[i] = w.Name
	}

	result.Winners = names

	election.Result = result

	election, err = r.store.EditElection(election)
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = r.store.CloseElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) Exclude(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	email := c.FormValue("email")

	voter, err := r.store.FindVoter(email)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for _, v := range election.GetExcluded() {
		if v.GetEmail() == voter.GetEmail() {
			return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
		}
	}

	election.Excluded = append(election.GetExcluded(), voter)

	_, err = r.store.EditElection(election)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) Include(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	election, err := r.store.FindElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	email := c.Param("email")

	voter, err := r.store.FindVoter(email)
	if err != nil {
		return r.errorHandle(c, err)
	}

	for index, v := range election.GetExcluded() {
		if v.GetEmail() == voter.GetEmail() {
			copy(election.GetExcluded()[index:], election.GetExcluded()[index+1:])     // Shift a[i+1:] left one index
			election.Excluded[len(election.GetExcluded())-1] = nil                     // Erase last element (write zero value)
			election.Excluded = election.GetExcluded()[:len(election.GetExcluded())-1] // Truncate slice

			_, err = r.store.EditElection(election)
			if err != nil {
				return r.errorHandle(c, err)
			}
			break
		}
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/admin/election/%s", election.GetId()))
}

func (r *AdminRepo) DeleteElection(c echo.Context) error {
	id := c.Param("id")
	if len(id) == 0 {
		return r.errorHandle(c, fmt.Errorf("invalid election id"))
	}

	err := r.store.DeleteElection(id)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return c.Redirect(http.StatusFound, "/admin/elections")
}

func (r *AdminRepo) Voters(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}
	voters := stv.GetVoters()

	var err1 string
	if len(c.FormValue("error")) > 0 {
		err1 = c.FormValue("error")
	}
	data := struct {
		Voters            []*storage.Voter
		AllowRegistration bool
		Error             string
	}{
		Voters:            voters,
		AllowRegistration: stv.GetAllowRegistration(),
		Error:             err1,
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.VotersTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}
	return nil
}

func (r *AdminRepo) AddVoter(c echo.Context) error {
	email := c.FormValue("email")
	name := c.FormValue("name")
	if len(name) == 0 || len(email) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and email need to be filled"))
	}
	_, err := r.store.AddVoter(&storage.Voter{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return r.errorHandle(c, err)
	}
	return r.Voters(c)
}

func (r *AdminRepo) DeleteVoter(c echo.Context) error {
	email := c.FormValue("email")
	err := r.store.DeleteVoter(email)
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

func (r *AdminRepo) ForceReset(c echo.Context) error {
	var err error
	err = r.store.DeleteAllElections()
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = r.store.DeleteAllVoters()
	if err != nil {
		return r.errorHandle(c, err)
	}

	_, err = r.store.SetAllowRegistration(false)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return c.JSON(http.StatusOK, "{\"message\": \"successfully reset stored data\"}")
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
