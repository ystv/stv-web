package store

import (
	"bytes"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/ystv/stv-web/storage"
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

func (store *Store) GetSessionHash(hash []byte, stv *storage.STV) (*storage.Session, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}

	if len(hash) != 32 {
		return nil, fmt.Errorf("invalid length of hash, need 32, got: %v", len(hash))
	}

	for _, session := range stv.GetSessions() {
		if bytes.Equal(session.Hash, hash) {
			return session, nil
		}
	}
	return nil, nil
}

func (store *Store) AddSession(hash []byte, name string, email string, accessToken string, stv *storage.STV) (*storage.Session, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}

	check, err := store.GetSessionHash(hash, stv)
	if err != nil {
		return nil, err
	}
	if check != nil {
		return nil, fmt.Errorf("cannot register two sessions for the same profile")
	}

	session := storage.Session{
		Hash:       hash,
		Name:       name,
		Email:      email,
		Registered: false,
		Access:     accessToken,
	}

	stv.Sessions = append(stv.GetSessions(), &session)
	if err := store.backend.Write(stv); err != nil {
		return nil, err
	}
	return &session, nil
}

func (store *Store) DeleteSession(hash []byte, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}
	var index = -1
	for i, s := range stv.Sessions {
		if bytes.Equal(s.Hash, hash) {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("could not find hash: %v", string(hash))
	}
	copy(stv.Sessions[index:], stv.Sessions[index+1:])
	stv.Sessions[len(stv.Sessions)-1] = nil
	stv.Sessions = stv.Sessions[:len(stv.Sessions)-1]
	return store.backend.Write(stv)
}

func (store *Store) DeleteAllSessions(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}

	stv.Sessions = []*storage.Session{}
	return store.backend.Write(stv)
}

func (store *Store) GetBallotsElectionID(id string, stv *storage.STV) ([]*storage.Ballot, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	var ballots []*storage.Ballot
	for _, ballot := range stv.GetBallots() {
		if ballot.GetElection() == id {
			ballots = append(ballots, ballot)
		}
	}
	return ballots, nil
}

func (store *Store) AddBallot(ballot *storage.Ballot, stv *storage.STV) (*storage.Ballot, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
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

func (store *Store) EditBallot(ballot *storage.Ballot, stv *storage.STV) (*storage.Ballot, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
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

func (store *Store) DeleteBallot(id string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
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

func (store *Store) DeleteAllBallots(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return err
		}
	}

	stv.Ballots = []*storage.Ballot{}

	return store.backend.Write(stv)
}

func (store *Store) GetCandidatesElectionID(id string, stv *storage.STV) ([]*storage.Candidate, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	var candidates []*storage.Candidate
	for _, candidate := range stv.GetCandidates() {
		if candidate.GetElection() == id {
			candidates = append(candidates, candidate)
		}
	}
	return candidates, nil
}

func (store *Store) FindCandidate(id string, stv *storage.STV) (*storage.Candidate, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	for _, c1 := range stv.GetCandidates() {
		if c1.GetId() == id {
			return c1, nil
		}
	}
	return nil, fmt.Errorf("unable to find candidate for FindCandidate: %s", id)
}

func (store *Store) AddCandidate(candidate *storage.Candidate, stv *storage.STV) (*storage.Candidate, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
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

func (store *Store) DeleteCandidate(id string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return err
		}
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

func (store *Store) DeleteAllCandidates(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return err
		}
	}

	stv.Candidates = []*storage.Candidate{}

	return store.backend.Write(stv)
}

func (store *Store) GetElections(stv *storage.STV) ([]*storage.Election, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	return stv.GetElections(), err
}

func (store *Store) FindElection(id string, stv *storage.STV) (*storage.Election, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	for _, e1 := range stv.GetElections() {
		if e1.GetId() == id {
			return e1, nil
		}
	}
	return nil, fmt.Errorf("unable to find election for FindElection")
}

func (store *Store) AddElection(election *storage.Election, stv *storage.STV) (*storage.Election, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
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

func (store *Store) EditElection(election *storage.Election, stv *storage.STV) (*storage.Election, error) {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return nil, err
		}
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

func (store *Store) OpenElection(id string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return err
		}
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

func (store *Store) CloseElection(id string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.backend.Read()
		if err != nil {
			return err
		}
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

func (store *Store) DeleteElection(id string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
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
			ballots, err = store.GetBallotsElectionID(id, stv)
			if err != nil {
				return err
			}
			for _, b1 := range ballots {
				err = store.DeleteBallot(b1.GetId(), stv)
				if err != nil {
					return err
				}
			}

			var candidates []*storage.Candidate
			candidates, err = store.GetCandidatesElectionID(id, stv)
			if err != nil {
				return err
			}
			for _, c1 := range candidates {
				err = store.DeleteCandidate(c1.GetId(), stv)
				if err != nil {
					return err
				}
			}

			var urls []*storage.URL
			urls, err = store.GetURLsElectionID(id, stv)
			if err != nil {
				return err
			}

			for _, u1 := range urls {
				err = store.DeleteURL(u1.GetUrl(), stv)
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

func (store *Store) DeleteAllElections(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}

	stv.Elections = []*storage.Election{}
	stv.Candidates = []*storage.Candidate{}
	stv.Ballots = []*storage.Ballot{}
	stv.Urls = []*storage.URL{}

	return store.backend.Write(stv)
}

func (store *Store) GetURLsElectionID(id string, stv *storage.STV) ([]*storage.URL, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	var urls []*storage.URL
	for _, url := range stv.GetUrls() {
		if url.GetElection() == id {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

func (store *Store) FindURL(url string, stv *storage.STV) (*storage.URL, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	for _, u1 := range stv.GetUrls() {
		if u1.GetUrl() == url {
			return u1, nil
		}
	}
	return nil, fmt.Errorf("unable to find url for FindURL")
}

func (store *Store) AddURL(url *storage.URL, stv *storage.STV) (*storage.URL, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
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

func (store *Store) SetURLVoted(url string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
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

func (store *Store) DeleteURL(url string, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
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

func (store *Store) DeleteAllURLs(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}

	stv.Urls = []*storage.URL{}

	return store.backend.Write(stv)
}

func (store *Store) GetAllowRegistration(stv *storage.STV) (bool, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return false, err
		}
	}
	return stv.GetAllowRegistration(), err
}

func (store *Store) SetAllowRegistration(allow bool, stv *storage.STV) (bool, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return false, err
		}
	}

	stv.AllowRegistration = allow

	if err = store.backend.Write(stv); err != nil {
		return false, err
	}

	return allow, nil
}

func (store *Store) GetVoters(stv *storage.STV) ([]*storage.Session, error) {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return nil, err
		}
	}
	sessions := []*storage.Session{}
	for _, session := range stv.GetSessions() {
		if session.GetRegistered() {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (store *Store) AddVoter(voter []byte, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}
	session, err := store.GetSessionHash(voter, stv)
	if err != nil {
		return err
	}

	voters, err := store.GetVoters(stv)
	if err != nil {
		return err
	}

	for _, v := range voters {
		if v.GetEmail() == session.GetEmail() {
			return fmt.Errorf("unable to add voter duplicate email for AddVoter")
		}
	}
	session.Registered = true

	if err = store.backend.Write(stv); err != nil {
		return err
	}

	return nil
}

func (store *Store) DeleteVoter(voter []byte, stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}

	session, err := store.GetSessionHash(voter, stv)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("no user with that hash")
	}

	session.Registered = false

	return store.backend.Write(stv)
}

func (store *Store) DeleteAllVoters(stv *storage.STV) error {
	var err error
	if stv == nil {
		stv, err = store.Get()
		if err != nil {
			return err
		}
	}

	voters, err := store.GetVoters(stv)
	if err != nil {
		return err
	}

	for _, session := range voters {
		session.Registered = false
	}

	return store.backend.Write(stv)
}

func (store *Store) Get() (*storage.STV, error) {
	return store.backend.Read()
}
