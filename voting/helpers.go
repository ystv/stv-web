package voting

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/bndr/gotabulate"
)

const float64EqualityThreshold = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

type CandidateStatus string

const (
	Elected  = "Elected"
	Hopeful  = "Hopeful"
	Rejected = "Rejected"
)

type CandidateResult struct {
	Candidate     *Candidate
	NumberOfVotes float64
	Status        CandidateStatus
}

type RoundResult struct {
	CandidateResults   []*CandidateResult
	NumberOfBlankVotes float64
}

func NewRoundResult(candidateResults []*CandidateResult, nbBlankVotes float64) *RoundResult {
	return &RoundResult{
		CandidateResults:   candidateResults,
		NumberOfBlankVotes: nbBlankVotes,
	}
}

func (rr *RoundResult) String() string {
	// Checking if we need to include Blank votes as a result
	resultsWithBlankVotes := rr.CandidateResults
	if !almostEqual(rr.NumberOfBlankVotes, 0.0) {
		resultsWithBlankVotes = append(resultsWithBlankVotes,
			&CandidateResult{
				Candidate:     NewCandidate("BlankVotes"),
				NumberOfVotes: rr.NumberOfBlankVotes,
				Status:        Rejected},
		)
	}

	// Prepares the rows
	rows := make([][]interface{}, 0)
	for _, result := range resultsWithBlankVotes {
		row := []interface{}{result.Candidate, result.NumberOfVotes, result.Status}
		rows = append(rows, row)
	}

	t := gotabulate.Create(rows)
	t.SetHeaders([]string{"Candidate", "# Votes", "Status"})
	t.SetAlign("center")

	return t.Render("grid")
}

type CandidateVoteCount struct {
	Candidate     *Candidate
	Status        CandidateStatus
	NumberOfVotes float64
	Votes         []*Ballot
}

func NewCandidateVoteCount(candidate *Candidate) *CandidateVoteCount {
	return &CandidateVoteCount{
		Candidate:     candidate,
		Status:        Hopeful,
		NumberOfVotes: 0.0,
		Votes:         make([]*Ballot, 0),
	}
}

func (cvc *CandidateVoteCount) IsInRace() bool {
	return cvc.Status == Hopeful
}

func (cvc *CandidateVoteCount) GetCandidateResult() *CandidateResult {
	return &CandidateResult{
		Candidate:     cvc.Candidate,
		NumberOfVotes: cvc.NumberOfVotes,
		Status:        cvc.Status,
	}
}

func (cvc *CandidateVoteCount) String() string {
	return fmt.Sprintf("CandidateVoteCount(candidate=%s, votes=%.2f)", cvc.Candidate, cvc.NumberOfVotes)
}

type CompareMethod string

const (
	CompareMethodMostSecondChoice = "MostSecondChoice"
	CompareMethodRandom           = "Random"
)

type ElectionManager struct {
	Candidates []*Candidate
	Ballots    []*Ballot
	ElectionManagerOptions

	CandidateVoteCounts map[*Candidate]*CandidateVoteCount
	CandidatesInRace    []*CandidateVoteCount
	CandidatesElected   []*CandidateVoteCount
	CandidatesRejected  []*CandidateVoteCount
	ExhaustedBallots    []*Ballot
	NumberOfBlankVotes  float64
	NumberOfCandidates  int
}

type ElectionManagerOptions struct {
	NumberOfVotesPerVoter int
	CompareMethodIfEqual  CompareMethod
	PickRandomIfBlank     bool
}

func DefaultElectionManagerOptions() ElectionManagerOptions {
	return ElectionManagerOptions{
		NumberOfVotesPerVoter: 1,
		CompareMethodIfEqual:  CompareMethodMostSecondChoice,
		PickRandomIfBlank:     false,
	}
}

func NewElectionManager(candidates []*Candidate, ballots []*Ballot, options ElectionManagerOptions) *ElectionManager {
	candidateVoteCounts := make(map[*Candidate]*CandidateVoteCount)
	for _, candidate := range candidates {
		candidateVoteCounts[candidate] = NewCandidateVoteCount(candidate)
	}

	candidatesInRace := make([]*CandidateVoteCount, 0)
	for _, cvc := range candidateVoteCounts {
		candidatesInRace = append(candidatesInRace, cvc)
	}

	exhaustedBallots := make([]*Ballot, 0)
	nbBlankVotes := 0.0

	for _, ballot := range ballots {
		candidatesThatShouldBeVotedOn := ballot.RankedCandidates[0:options.NumberOfVotesPerVoter]
		numberOfBlankVotes := options.NumberOfVotesPerVoter - len(ballot.RankedCandidates)

		if numberOfBlankVotes > 0 {
			if options.PickRandomIfBlank {
				for i := 0; i < numberOfBlankVotes; i++ {
					nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidates))))
					if err != nil {
						panic(err)
					}
					newCandidateChoice := candidates[int(nBig.Int64())]
					candidatesThatShouldBeVotedOn = append(candidatesThatShouldBeVotedOn, newCandidateChoice)
				}
			} else {
				exhaustedBallots = append(exhaustedBallots, ballot)
				nbBlankVotes += float64(numberOfBlankVotes)
			}
		}

		for _, candidate := range candidatesThatShouldBeVotedOn {
			candidateVoteCounts[candidate].NumberOfVotes++
			candidateVoteCounts[candidate].Votes = append(candidateVoteCounts[candidate].Votes, ballot)
		}
	}

	electionManager := &ElectionManager{
		Candidates:             candidates,
		Ballots:                ballots,
		ElectionManagerOptions: options,

		CandidateVoteCounts: candidateVoteCounts,
		CandidatesInRace:    candidatesInRace,
		CandidatesElected:   make([]*CandidateVoteCount, 0),
		CandidatesRejected:  make([]*CandidateVoteCount, 0),
		ExhaustedBallots:    exhaustedBallots,
		NumberOfBlankVotes:  nbBlankVotes,
		NumberOfCandidates:  len(candidates),
	}

	electionManager.SortCandidatesInRace()

	return electionManager
}

func (em *ElectionManager) SortCandidatesInRace() {
	sort.SliceStable(em.CandidatesInRace, func(i, j int) bool {
		c1Votes := em.CandidatesInRace[i].NumberOfVotes
		c2Votes := em.CandidatesInRace[j].NumberOfVotes

		if !almostEqual(c1Votes, c2Votes) {
			return c1Votes > c2Votes
		}

		if em.CompareMethodIfEqual == CompareMethodMostSecondChoice {
			return em.Candidate1HasMostSecondChoices(em.CandidatesInRace[i], em.CandidatesInRace[j], 1)
		}
		// Random method by default if not second choice
		// requires to be input of 2 as it doesn't include upper bound
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(2)))
		if err != nil {
			panic(err)
		}
		return nBig.Int64() == 0
	})
}

func (em *ElectionManager) IsValidCandidate(candidate *Candidate) bool {
	_, ok := em.CandidateVoteCounts[candidate]
	return ok
}

func removeCandidateFromSlice(slice []*CandidateVoteCount, candidateCV *CandidateVoteCount) ([]*CandidateVoteCount, error) {
	for k, c := range slice {
		if c.Candidate.Equals(candidateCV.Candidate) {
			slice = append(slice[:k], slice[k+1:]...)
			return slice, nil
		}
	}
	return nil, fmt.Errorf("candidate not found")
}

func (em *ElectionManager) ElectCandidate(candidate *Candidate) error {
	if !em.IsValidCandidate(candidate) {
		return fmt.Errorf("candidate not found in election manager")
	}

	candidateCV := em.CandidateVoteCounts[candidate]
	candidateCV.Status = Elected
	em.CandidatesElected = append(em.CandidatesElected, candidateCV)
	cir, err := removeCandidateFromSlice(em.CandidatesInRace, candidateCV)
	if err != nil {
		return err
	}
	em.CandidatesInRace = cir
	return nil
}

func (em *ElectionManager) RejectCandidate(candidate *Candidate) error {
	if !em.IsValidCandidate(candidate) {
		return fmt.Errorf("candidate not found in election manager")
	}

	candidateCV := em.CandidateVoteCounts[candidate]
	candidateCV.Status = Rejected
	em.CandidatesRejected = append(em.CandidatesRejected, candidateCV)
	cir, err := removeCandidateFromSlice(em.CandidatesInRace, candidateCV)
	if err != nil {
		return err
	}
	em.CandidatesInRace = cir
	return nil
}

func (em *ElectionManager) TransferVotes(candidate *Candidate, numberOfTransferVotes float64) error {
	if !em.IsValidCandidate(candidate) {
		return fmt.Errorf("candidate not found in election manager")
	}
	if math.Floor(numberOfTransferVotes*10_000)/10_000 == 0.000 {
		return nil
	}

	candidateCV := em.CandidateVoteCounts[candidate]
	if candidateCV.Status == Hopeful {
		return fmt.Errorf("election manager cannot transfer votes from a candidate that is still in the race (hopeful)")
	}

	voters := len(candidateCV.Votes)
	votesPerVoter := numberOfTransferVotes / float64(voters)

	for _, ballot := range candidateCV.Votes {
		newCandidateChoice := em.GetBallotCandidateNrXInRaceOrNone(ballot, em.NumberOfVotesPerVoter-1)

		if newCandidateChoice == nil && em.PickRandomIfBlank {
			candidatesInRace := em.GetCandidatesInRace()
			if len(candidatesInRace) > 0 {
				nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidatesInRace))))
				if err != nil {
					return fmt.Errorf("failed to generate random new candidate choice: %w", err)
				}
				newCandidateChoice = candidatesInRace[int(nBig.Int64())]
			}
		}

		if newCandidateChoice != nil {
			em.CandidateVoteCounts[newCandidateChoice].NumberOfVotes += votesPerVoter
			em.CandidateVoteCounts[newCandidateChoice].Votes = append(em.CandidateVoteCounts[newCandidateChoice].Votes, ballot)
		} else {
			em.ExhaustedBallots = append(em.ExhaustedBallots, ballot)
			em.NumberOfBlankVotes += votesPerVoter
		}
	}

	candidateCV.NumberOfVotes -= numberOfTransferVotes
	candidateCV.Votes = []*Ballot{}

	em.SortCandidatesInRace()

	return nil
}

func (em *ElectionManager) GetBallotCandidateNrXInRaceOrNone(ballot *Ballot, x int) *Candidate {
	rankedCandidatesInRace := make([]*Candidate, 0)
	for _, c := range ballot.RankedCandidates {
		if em.CandidateVoteCounts[c].IsInRace() {
			rankedCandidatesInRace = append(rankedCandidatesInRace, c)
		}
	}

	if len(rankedCandidatesInRace) > x {
		return rankedCandidatesInRace[x]
	}

	return nil
}

func (em *ElectionManager) GetCandidatesInRace() []*Candidate {
	candidatesInRace := make([]*Candidate, 0)
	for _, cvc := range em.CandidatesInRace {
		if cvc.IsInRace() {
			candidatesInRace = append(candidatesInRace, cvc.Candidate)
		}
	}
	return candidatesInRace
}

func (em *ElectionManager) GetNumberOfNonExhaustedVotes() float64 {
	return float64(len(em.Ballots)*em.NumberOfVotesPerVoter) - em.NumberOfBlankVotes
}

func (em *ElectionManager) GetNumberOfNonExhaustedBallots() float64 {
	return float64(len(em.Ballots) - len(em.ExhaustedBallots))
}

func (em *ElectionManager) GetNumberOfCandidatesInRace() uint64 {
	return uint64(len(em.GetCandidatesInRace()))
}

func (em *ElectionManager) GetNumberOfElectedCandidates() uint64 {
	return uint64(len(em.CandidatesElected))
}

func (em *ElectionManager) GetNumberOfVotes(candidate *Candidate) (float64, error) {
	if !em.IsValidCandidate(candidate) {
		return 0.0, fmt.Errorf("candidate not found in election manager")
	}
	return em.CandidateVoteCounts[candidate].NumberOfVotes, nil
}

func (em *ElectionManager) GetCandidateWithLeastVotesInRace() (*Candidate, error) {
	if len(em.CandidatesInRace) == 0 {
		return nil, fmt.Errorf("not candidates left in race")
	}
	return em.CandidatesInRace[len(em.CandidatesInRace)-1].Candidate, nil
}

func (em *ElectionManager) GetCandidatesWithMoreThanXVotes(x int) []*Candidate {
	candidates := make([]*Candidate, 0)
	for _, cvc := range em.CandidatesInRace {
		if math.Floor(cvc.NumberOfVotes*10_000)/10_000 > float64(x*10_000)/10_000 {
			candidates = append(candidates, cvc.Candidate)
		}
	}
	return candidates
}

func (em *ElectionManager) GetResults() *RoundResult {
	candidatesVC := make([]*CandidateVoteCount, 0)
	candidatesVC = append(candidatesVC, em.CandidatesElected...)
	candidatesVC = append(candidatesVC, em.CandidatesInRace...)
	candidatesVC = append(candidatesVC, em.CandidatesRejected...)

	candidateResults := make([]*CandidateResult, 0)
	for _, c := range candidatesVC {
		candidateResults = append(candidateResults, c.GetCandidateResult())
	}

	return NewRoundResult(candidateResults, em.NumberOfBlankVotes)
}

func (em *ElectionManager) Candidate1HasMostSecondChoices(c1vc, c2vc *CandidateVoteCount, x int) bool {
	if x >= em.NumberOfCandidates {
		// requires to be input of 2 as it doesn't include upper bound
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(2)))
		if err != nil {
			panic(err)
		}
		return nBig.Uint64() == 1
	}

	votesCandidate1 := 0
	votesCandidate2 := 0

	for _, ballot := range em.Ballots {
		candidate := em.GetBallotCandidateNrXInRaceOrNone(ballot, x)

		switch candidate {
		case c1vc.Candidate:
			votesCandidate1++
		case c2vc.Candidate:
			votesCandidate2++
		default:
			continue
		}
	}

	if votesCandidate1 == votesCandidate2 {
		return em.Candidate1HasMostSecondChoices(c1vc, c2vc, x+1)
	}
	return votesCandidate1 > votesCandidate2
}

type ElectionResults struct {
	Rounds []*RoundResult
}

func NewElectionResults() *ElectionResults {
	return &ElectionResults{
		Rounds: make([]*RoundResult, 0),
	}
}

func (er *ElectionResults) RegisterResults(round *RoundResult) {
	er.Rounds = append(er.Rounds, round)
}

func (er *ElectionResults) GetWinners() []*Candidate {
	if len(er.Rounds) == 0 {
		return []*Candidate{}
	}

	lastRound := er.Rounds[len(er.Rounds)-1]
	winnerCandidates := make([]*Candidate, 0)
	for _, candidateResult := range lastRound.CandidateResults {
		if candidateResult.Status == Elected {
			winnerCandidates = append(winnerCandidates, candidateResult.Candidate)
		}
	}

	return winnerCandidates
}
