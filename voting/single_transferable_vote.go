package voting

import "fmt"

type SingleTransferableVoteOptions struct {
	CompareMethodIfEquals CompareMethod
	PickRandomIfBlank     bool
}

func DefaultSingleTransferableVoteOptions() SingleTransferableVoteOptions {
	return SingleTransferableVoteOptions{
		CompareMethodIfEquals: CompareMethodMostSecondChoice,
		PickRandomIfBlank:     false,
	}
}

func SingleTransferableVote(candidates []*Candidate, ballots []*Ballot, numberOfSeats uint64, options SingleTransferableVoteOptions) (*ElectionResults, error) {
	roundingError := 1e-6
	manager := NewElectionManager(candidates, ballots, ElectionManagerOptions{
		NumberOfVotesPerVoter: 1,
		CompareMethodIfEqual:  options.CompareMethodIfEquals,
		PickRandomIfBlank:     options.PickRandomIfBlank,
	})
	electionResults := NewElectionResults()

	voters := manager.GetNumberOfNonExhaustedBallots()
	seats := numberOfSeats
	votesNeededToWin := voters / float64(seats+1) // Droop Quota

	for {
		seatsLeft := numberOfSeats - manager.GetNumberOfElectedCandidates()
		candidatesInRace := manager.GetCandidatesInRace()
		candidatesInRaceVotes := make([]float64, 0)
		for _, c := range candidatesInRace {
			votes, err := manager.GetNumberOfVotes(c)
			if err != nil {
				return nil, err
			}
			candidatesInRaceVotes = append(candidatesInRaceVotes, votes)
		}
		votesRemaining := sumSlice(candidatesInRaceVotes)
		lastVotes := 0.0
		candidatesToElect := make([]*Candidate, 0)
		candidatesToReject := make([]*Candidate, 0)

	candidatesInRaceLoop:
		for i, candidate := range candidatesInRace {
			j := uint64(i)
			votesForCandidate := candidatesInRaceVotes[j]
			isLastCandidate := j == uint64(len(candidatesInRace)-1)

			switch {
			case votesForCandidate-roundingError >= votesNeededToWin:
				candidatesToElect = append(candidatesToElect, candidate)
			case j >= seatsLeft && (votesRemaining-roundingError) <= lastVotes:
				if len(candidatesToElect) > 0 {
					break candidatesInRaceLoop
				}
				candidatesToReject = append(candidatesToReject, candidate)
			case isLastCandidate:
				return nil, fmt.Errorf("election ended up in an illegal state")
			}

			lastVotes = votesForCandidate
			votesRemaining -= votesForCandidate
		}

		for _, candidate := range candidatesToElect {
			if err := manager.ElectCandidate(candidate); err != nil {
				return nil, err
			}
		}

		for i := len(candidatesToReject) - 1; i >= 0; i-- {
			if err := manager.RejectCandidate(candidatesToReject[i]); err != nil {
				return nil, err
			}
		}

		seatsLeft = numberOfSeats - manager.GetNumberOfElectedCandidates()
		if manager.GetNumberOfCandidatesInRace() <= seatsLeft {
			for _, candidate := range manager.GetCandidatesInRace() {
				candidatesToElect = append(candidatesToElect, candidate)
				if err := manager.ElectCandidate(candidate); err != nil {
					return nil, err
				}
			}
		}

		seatsLeft = numberOfSeats - manager.GetNumberOfElectedCandidates()
		if seatsLeft == 0 {
			candidatesInRace = manager.GetCandidatesInRace()
			for i := len(candidatesInRace) - 1; i >= 0; i-- {
				candidatesToReject = append(candidatesToReject, candidatesInRace[i])
				if err := manager.RejectCandidate(candidatesInRace[i]); err != nil {
					return nil, err
				}
			}
		}

		electionResults.RegisterResults(manager.GetResults())

		if manager.GetNumberOfCandidatesInRace() == 0 {
			break
		}
		for _, c := range candidatesToElect {
			votesForCandidate, err := manager.GetNumberOfVotes(c)
			if err != nil {
				return nil, err
			}
			excessVotes := votesForCandidate - votesNeededToWin
			if err = manager.TransferVotes(c, excessVotes); err != nil {
				return nil, err
			}
		}

		for _, c := range candidatesToReject {
			votesForCandidate, err := manager.GetNumberOfVotes(c)
			if err != nil {
				return nil, err
			}
			if err = manager.TransferVotes(c, votesForCandidate); err != nil {
				return nil, err
			}
		}
	}
	return electionResults, nil
}

func sumSlice(n []float64) float64 {
	s := 0.0
	for _, num := range n {
		s += num
	}
	return s
}
