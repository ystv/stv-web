package store

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/ystv/stv_web/storage"
)

type Store struct {
	backend Backend
}

func NewStore(root bool) (*Store, error) {
	backend, err := NewFileBackend(root)
	if err != nil {
		return nil, err
	}
	return &Store{backend: backend}, nil
}

func (store *Store) GetBallotsElectionID(id string) ([]*storage.Ballot, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var ballots []*storage.Ballot
	for _, ballot := range stv.GetBallots() {
		if ballot.GetElection() == id {
			ballots = append(ballots, ballot)
		}
	}
	return ballots, nil
}

func (store *Store) AddBallot(ballot *storage.Ballot) (*storage.Ballot, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}

beforeUUID:
	ballot.Id = uuid.NewString()

	for _, b := range stv.GetBallots() {
		if b.GetId() == ballot.GetId() {
			log.Println("duplicate ballot id, retrying...")
			goto beforeUUID
		}
	}

	for _, election := range stv.GetElections() {
		if election.GetId() == ballot.GetElection() {
			stv.Ballots = append(stv.GetBallots(), ballot)

			if err = store.backend.Write(stv); err != nil {
				return nil, err
			}

			return ballot, nil
		}
	}

	return nil, fmt.Errorf("unable to find election fot AddBallot")
}

func (store *Store) EditBallot(ballot *storage.Ballot) (*storage.Ballot, error) {
	stv, err := store.backend.Read()
	if err != nil {
		return nil, err
	}

	for _, b := range stv.GetBallots() {
		if b.GetId() == ballot.GetId() {
			b.Choice = ballot.GetChoice()
			if err = store.backend.Write(stv); err != nil {
				return nil, err
			}
			return b, nil
		}
	}
	return nil, fmt.Errorf("unable to find ballot for EditBallot")
}

func (store *Store) DeleteBallot(id string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.GetBallots()
	found := false
	var index int
	var ballots *storage.Ballot
	for index, ballots = range s {
		if ballots.GetId() == id {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Ballots = s[:len(s)-1]   // Truncate slice

		return store.backend.Write(stv)
	}
	return fmt.Errorf("ballot not found for DeleteBallot")
}

func (store *Store) GetCandidatesElectionID(id string) ([]*storage.Candidate, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var candidates []*storage.Candidate
	for _, candidate := range stv.GetCandidates() {
		if candidate.GetElection() == id {
			candidates = append(candidates, candidate)
		}
	}
	return candidates, nil
}

func (store *Store) FindCandidate(id string) (*storage.Candidate, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, c1 := range stv.GetCandidates() {
		if c1.GetId() == id {
			return c1, nil
		}
	}
	return nil, fmt.Errorf("unable to find candidate for FindCandidate: %s", id)
}

func (store *Store) AddCandidate(candidate *storage.Candidate) (*storage.Candidate, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}

beforeUUID:
	candidate.Id = uuid.NewString()

	for _, c := range stv.GetCandidates() {
		if c.GetId() == candidate.GetId() {
			log.Println("duplicate candidate id, retrying...")
			goto beforeUUID
		}
	}

	stv.Candidates = append(stv.GetCandidates(), candidate)

	if err = store.backend.Write(stv); err != nil {
		return nil, err
	}

	return candidate, nil
}

func (store *Store) DeleteCandidate(id string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.GetCandidates()
	found := false
	var index int
	var candidate *storage.Candidate
	for index, candidate = range s {
		if candidate.GetId() == id {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:])  // Shift a[i+1:] left one index
		s[len(s)-1] = nil             // Erase last element (write zero value)
		stv.Candidates = s[:len(s)-1] // Truncate slice

		return store.backend.Write(stv)
	}
	return fmt.Errorf("candidate not found for DeleteCandidate")
}

func (store *Store) GetElections() ([]*storage.Election, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	return stv.GetElections(), err
}

func (store *Store) FindElection(id string) (*storage.Election, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, e1 := range stv.GetElections() {
		if e1.GetId() == id {
			return e1, nil
		}
	}
	return nil, fmt.Errorf("unable to find election for FindElection")
}

func (store *Store) AddElection(election *storage.Election) (*storage.Election, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}

beforeUUID:
	election.Id = uuid.NewString()

	for _, e := range stv.GetBallots() {
		if e.GetId() == election.GetId() {
			log.Println("duplicate election id, retrying...")
			goto beforeUUID
		}
	}

	election.Open = false
	election.Closed = false

	stv.Elections = append(stv.GetElections(), election)

	if err = store.backend.Write(stv); err != nil {
		return nil, err
	}

	return election, nil
}

func (store *Store) EditElection(election *storage.Election) (*storage.Election, error) {
	stv, err := store.backend.Read()
	if err != nil {
		return nil, err
	}

	for _, e := range stv.GetElections() {
		if e.GetId() == election.GetId() {
			e.Name = election.GetName()
			e.Description = election.GetDescription()
			e.Ron = election.GetRon()
			e.Seats = election.GetSeats()
			e.Open = election.GetOpen()
			e.Closed = election.GetClosed()
			e.Result = election.GetResult()
			e.Excluded = election.GetExcluded()
			if err = store.backend.Write(stv); err != nil {
				return nil, err
			}
			return e, nil
		}
	}
	return nil, fmt.Errorf("election not found for EditElection")
}

func (store *Store) OpenElection(id string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	for _, e1 := range stv.GetElections() {
		if e1.GetId() == id {
			if e1.GetOpen() {
				return fmt.Errorf("already set opened for OpenElection")
			}
			if e1.GetClosed() {
				return fmt.Errorf("election closed for OpenElection")
			}
			e1.Open = true
			return store.backend.Write(stv)
		}
	}
	return nil
}

func (store *Store) CloseElection(id string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	for _, e1 := range stv.GetElections() {
		if e1.GetId() == id {
			if e1.GetClosed() {
				return fmt.Errorf("already set closed for CloseElection")
			}
			e1.Closed = true
			e1.Open = false
			return store.backend.Write(stv)
		}
	}
	return nil
}

func (store *Store) DeleteElection(id string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.GetElections()
	found := false
	var index int
	var election *storage.Election
	for index, election = range s {
		if election.GetId() == id {
			if election.GetOpen() {
				return fmt.Errorf("cannot delete open election for DeleteElection")
			}
			found = true

			var ballots []*storage.Ballot
			ballots, err = store.GetBallotsElectionID(id)
			if err != nil {
				return err
			}
			for _, b1 := range ballots {
				err = store.DeleteBallot(b1.GetId())
				if err != nil {
					return err
				}
			}

			var candidates []*storage.Candidate
			candidates, err = store.GetCandidatesElectionID(id)
			if err != nil {
				return err
			}
			for _, c1 := range candidates {
				err = store.DeleteCandidate(c1.GetId())
				if err != nil {
					return err
				}
			}

			var urls []*storage.URL
			urls, err = store.GetURLsElectionID(id)
			if err != nil {
				return err
			}

			for _, u1 := range urls {
				err = store.DeleteURL(u1.GetUrl())
				if err != nil {
					return err
				}
			}
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Elections = s[:len(s)-1] // Truncate slice
	}

	return store.backend.Write(stv)
}

func (store *Store) DeleteAllElections() error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	stv.Elections = []*storage.Election{}
	stv.Candidates = []*storage.Candidate{}
	stv.Ballots = []*storage.Ballot{}
	stv.Urls = []*storage.URL{}

	return store.backend.Write(stv)
}

func (store *Store) GetURLsElectionID(id string) ([]*storage.URL, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var urls []*storage.URL
	for _, url := range stv.GetUrls() {
		if url.GetElection() == id {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func (store *Store) FindURL(url string) (*storage.URL, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, u1 := range stv.GetUrls() {
		if u1.GetUrl() == url {
			return u1, nil
		}
	}
	return nil, fmt.Errorf("unable to find url for FindURL")
}

func (store *Store) AddURL(url *storage.URL) (*storage.URL, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}

beforeUUID:
	url.Url = uuid.NewString()

	for _, u := range stv.GetUrls() {
		if u.GetUrl() == url.GetUrl() {
			log.Println("duplicate url, retrying...")
			goto beforeUUID
		}
	}

	for _, u := range stv.GetUrls() {
		if u.GetUrl() == url.GetUrl() {
			return nil, fmt.Errorf("unable to add url duplicate url for AddURL")
		}
	}

	url.Voted = false

	stv.Urls = append(stv.GetUrls(), url)

	if err = store.backend.Write(stv); err != nil {
		return nil, err
	}

	return url, nil
}

func (store *Store) SetURLVoted(url string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	for _, u1 := range stv.GetUrls() {
		if u1.GetUrl() == url {
			if u1.GetVoted() {
				return fmt.Errorf("already set voted for SetVoted")
			}
			u1.Voted = true
			return store.backend.Write(stv)
		}
	}
	return nil
}

func (store *Store) DeleteURL(url string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.GetUrls()
	found := false
	var index int
	var u *storage.URL
	for index, u = range s {
		if u.GetUrl() == url {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Urls = s[:len(s)-1]      // Truncate slice

		return store.backend.Write(stv)
	}
	return fmt.Errorf("url not found for DeleteURL")
}

func (store *Store) GetAllowRegistration() (bool, error) {
	stv, err := store.Get()
	if err != nil {
		return false, err
	}
	return stv.GetAllowRegistration(), err
}

func (store *Store) SetAllowRegistration(allow bool) (bool, error) {
	stv, err := store.Get()
	if err != nil {
		return false, err
	}

	stv.AllowRegistration = allow

	if err = store.backend.Write(stv); err != nil {
		return false, err
	}

	return allow, nil
}

func (store *Store) GetVoters() ([]*storage.Voter, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	return stv.GetVoters(), err
}

func (store *Store) FindVoter(email string) (*storage.Voter, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, v1 := range stv.GetVoters() {
		if v1.GetEmail() == email {
			return v1, nil
		}
	}
	return nil, fmt.Errorf("unable to find voter for FindVoter")
}

func (store *Store) AddVoter(voter *storage.Voter) (*storage.Voter, error) {
	stv, err := store.Get()
	if err != nil {
		return &storage.Voter{}, err
	}

	for _, v := range stv.GetVoters() {
		if v.GetEmail() == voter.GetEmail() {
			return &storage.Voter{}, fmt.Errorf("unable to add voter duplicate email for AddVoter")
		}
	}

	stv.Voters = append(stv.GetVoters(), voter)

	if err = store.backend.Write(stv); err != nil {
		return &storage.Voter{}, err
	}

	return voter, nil
}

func (store *Store) DeleteVoter(email string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	s := stv.GetVoters()
	found := false
	var index int
	var v *storage.Voter
	for index, v = range s {
		if v.GetEmail() == email {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Voters = s[:len(s)-1]    // Truncate slice

		return store.backend.Write(stv)
	}
	return fmt.Errorf("voter not found for DeleteVoter")
}

func (store *Store) DeleteAllVoters() error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	stv.Voters = []*storage.Voter{}

	return store.backend.Write(stv)
}

func (store *Store) Get() (*storage.STV, error) {
	return store.backend.Read()
}
