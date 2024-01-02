package store

import (
	"fmt"

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

func max(a, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}

func (store *Store) GetBallotsElectionId(id uint64) ([]*storage.Ballot, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var ballots []*storage.Ballot
	for _, ballot := range stv.Ballots {
		if ballot.Election == id {
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
	var id uint64
	id = 1
	for _, e := range stv.Ballots {
		id = max(id, e.Id)
		if e.Election == ballot.Election && e.Voter == ballot.Voter {
			return nil, fmt.Errorf("ballot already exists for AddBallot")
		}
	}

	ballot.Id = id

	for _, election := range stv.Elections {
		if election.Id == ballot.Election {
			for _, voter := range stv.Voters {
				if voter.Email == ballot.Voter {
					stv.Ballots = append(stv.Ballots, ballot)

					if err = store.backend.Write(stv); err != nil {
						return nil, err
					}

					return ballot, nil
				}
			}
			return nil, fmt.Errorf("unable to find voter for AddBallot")
		}
	}

	return nil, fmt.Errorf("unable to find election fot AddBallot")
}

func (store *Store) EditBallot(ballot *storage.Ballot) (*storage.Ballot, error) {
	stv, err := store.backend.Read()
	if err != nil {
		return nil, err
	}

	for _, b := range stv.Ballots {
		if b.Id == ballot.Id {
			b.Choice = ballot.Choice
			if err = store.backend.Write(stv); err != nil {
				return nil, err
			}
			return b, nil
		}
	}
	return nil, fmt.Errorf("unable to find ballot for EditBallot")
}

func (store *Store) DeleteBallot(id uint64) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.Ballots
	found := false
	var index int
	var ballots *storage.Ballot
	for index, ballots = range s {
		if ballots.Id == id {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Ballots = s[:len(s)-1]   // Truncate slice
	} else {
		return fmt.Errorf("ballot not found for DeleteBallot")
	}

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetCandidatesElectionId(id uint64) ([]*storage.Candidate, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var candidates []*storage.Candidate
	for _, candidate := range stv.Candidates {
		if candidate.Election == id {
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
	for _, c1 := range stv.Candidates {
		if c1.Id == id {
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

	for _, c := range stv.Candidates {
		if c.Id == candidate.Id {
			return nil, fmt.Errorf("unable to add candidate duplicate id for AddCandidate")
		}
	}

	stv.Candidates = append(stv.Candidates, candidate)

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
	s := stv.Candidates
	found := false
	var index int
	var candidate *storage.Candidate
	for index, candidate = range s {
		if candidate.Id == id {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:])  // Shift a[i+1:] left one index
		s[len(s)-1] = nil             // Erase last element (write zero value)
		stv.Candidates = s[:len(s)-1] // Truncate slice
	} else {
		return fmt.Errorf("candidate not found for DeleteCandidate")
	}

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetElections() ([]*storage.Election, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	return stv.Elections, err
}

func (store *Store) FindElection(id uint64) (*storage.Election, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, e1 := range stv.Elections {
		if e1.Id == id {
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
	var id uint64
	id = 1
	for _, e := range stv.Elections {
		id = max(id, e.Id)
	}

	election.Id = id + 1
	election.Open = false
	election.Closed = false

	stv.Elections = append(stv.Elections, election)

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

	for _, e := range stv.Elections {
		if e.Id == election.Id {
			e.Name = election.Name
			e.Description = election.Description
			e.Ron = election.Ron
			e.Open = election.Open
			e.Closed = election.Closed
			e.Result = election.Result
			e.Excluded = election.Excluded
			if err = store.backend.Write(stv); err != nil {
				return nil, err
			}
			return e, nil
		}
	}
	return nil, fmt.Errorf("election not found for EditElection")
}

func (store *Store) OpenElection(id uint64) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	for _, e1 := range stv.Elections {
		if e1.Id == id {
			if e1.Open {
				return fmt.Errorf("already set opened for OpenElection")
			}
			if e1.Closed {
				return fmt.Errorf("election closed for OpenElection")
			}
			e1.Open = true
			if err = store.backend.Write(stv); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (store *Store) CloseElection(id uint64) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	for _, e1 := range stv.Elections {
		if e1.Id == id {
			if e1.Closed {
				return fmt.Errorf("already set closed for CloseElection")
			}
			e1.Closed = true
			e1.Open = false
			if err = store.backend.Write(stv); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (store *Store) DeleteElection(id uint64) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.Elections
	found := false
	var index int
	var election *storage.Election
	for index, election = range s {
		if election.Id == id {
			if election.Open {
				return fmt.Errorf("cannot delete open election for DeleteElection")
			}
			found = true
			ballots, err := store.GetBallotsElectionId(id)
			if err != nil {
				return err
			}
			for _, b1 := range ballots {
				err = store.DeleteBallot(b1.Id)
				if err != nil {
					return err
				}
			}
			candidates, err := store.GetCandidatesElectionId(id)
			if err != nil {
				return err
			}
			for _, c1 := range candidates {
				err = store.DeleteCandidate(c1.Id)
				if err != nil {
					return err
				}
			}
			urls, err := store.GetURLsElectionId(id)
			if err != nil {
				return err
			}
			for _, u1 := range urls {
				err = store.DeleteURL(u1.Url)
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

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
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

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetURLsElectionId(id uint64) ([]*storage.URL, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	var urls []*storage.URL
	for _, url := range stv.Urls {
		if url.Election == id {
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
	for _, u1 := range stv.Urls {
		if u1.Url == url {
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

	for _, u := range stv.Urls {
		if u.Url == url.Url {
			return nil, fmt.Errorf("unable to add url duplicate url for AddURL")
		}
	}

	url.Voted = false

	stv.Urls = append(stv.Urls, url)

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

	for _, u1 := range stv.Urls {
		if u1.Url == url {
			if u1.Voted {
				return fmt.Errorf("already set voted for SetVoted")
			}
			u1.Voted = true
			if err := store.backend.Write(stv); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (store *Store) DeleteURL(url string) error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}
	s := stv.Urls
	found := false
	var index int
	var u *storage.URL
	for index, u = range s {
		if u.Url == url {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Urls = s[:len(s)-1]      // Truncate slice
	} else {
		return fmt.Errorf("url not found for DeleteURL")
	}

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) GetAllowRegistration() (bool, error) {
	stv, err := store.Get()
	if err != nil {
		return false, err
	}
	return stv.AllowRegistration, err
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
	return stv.Voters, err
}

func (store *Store) FindVoter(email string) (*storage.Voter, error) {
	stv, err := store.Get()
	if err != nil {
		return nil, err
	}
	for _, v1 := range stv.Voters {
		if v1.Email == email {
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

	for _, v := range stv.Voters {
		if v.Email == voter.Email {
			return &storage.Voter{}, fmt.Errorf("unable to add voter duplicate email for AddVoter")
		}
	}

	stv.Voters = append(stv.Voters, voter)

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

	s := stv.Voters
	found := false
	var index int
	var v *storage.Voter
	for index, v = range s {
		if v.Email == email {
			found = true
			break
		}
	}

	if found {
		copy(s[index:], s[index+1:]) // Shift a[i+1:] left one index
		s[len(s)-1] = nil            // Erase last element (write zero value)
		stv.Voters = s[:len(s)-1]    // Truncate slice
	} else {
		return fmt.Errorf("voter not found for DeleteVoter")
	}

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) DeleteAllVoters() error {
	stv, err := store.backend.Read()
	if err != nil {
		return err
	}

	stv.Voters = []*storage.Voter{}

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) Get() (*storage.STV, error) {
	return store.backend.Read()
}
