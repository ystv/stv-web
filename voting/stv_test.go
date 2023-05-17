package voting

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

//go:embed testdata/*.blt
var testdata embed.FS

func ParseBLT(tb testing.TB, file string) ([]*Candidate, []*Ballot, int) {
	tb.Helper()

	fd, err := testdata.Open(file)
	require.NoError(tb, err)
	defer func(fd fs.File) {
		_ = fd.Close()
	}(fd)

	sc := bufio.NewScanner(fd)
	candidates := make([]*Candidate, 0)
	ballots := make([]*Ballot, 0)
	var numCandidates, numSeats int

	/*
			Structure of a .blt file (from https://stackoverflow.com/a/2234236/2586553):
		    4 2          # four candidates are competing for two seats
		    -2           # Bob has withdrawn (optional)
		    1 4 1 3 2 0  # first ballot
		    1 2 4 1 3 0
		    1 1 4 2 3 0  # The first number is the ballot weight (>= 1).
		    1 1 2 4 3 0  # The last 0 is an end of ballot marker.
		    1 1 4 3 0    # Numbers in between correspond to the candidates
		    1 3 2 4 1 0  # on the ballot.
		    1 3 4 1 2 0
		    1 3 4 1 2 0  # Chuck, Diane, Amy, Bob
		    1 4 3 2 0
		    1 2 3 4 1 0  # last ballot
		    0            # end of ballots marker
		    "Amy"        # candidate 1
		    "Bob"        # candidate 2
		    "Chuck"      # candidate 3
		    "Diane"      # candidate 4
		    "Gardening Club Election"  # title
	*/

	require.True(tb, sc.Scan())
	_, err = fmt.Sscanf(sc.Text(), "%d %d", &numCandidates, &numSeats)
	require.NoError(tb, err)

	// The issue with the format is that the candidate names are at the end of the file, so we get a bit cheeky.
	// Create pointers to them now, and fill in the names at the end.
	candidatesByIndex := make(map[int]*Candidate)
	for i := 0; i < numCandidates; i++ {
		candidatesByIndex[i+1 /* one-indexed! */] = new(Candidate)
	}

	for sc.Scan() {
		line := sc.Text()
		if line == "0" {
			break
		}
		if line[0] == '-' {
			panic("withdrawals not supported")
		}
		weightStr, ballotStr, ok := strings.Cut(line, " ")
		require.Truef(tb, ok, "failed to parse ballot weight: %q", line)
		weight, err := strconv.Atoi(weightStr)
		require.NoErrorf(tb, err, "failed to parse ballot weight: %w", err)

		votes := strings.Split(ballotStr, " ")
		// we know the last element is 0, so we can just ignore it
		ballot := &Ballot{
			RankedCandidates: make([]*Candidate, 0, len(votes)-1),
		}
		for _, voteStr := range votes[:len(votes)-1] {
			candidateIndex, err := strconv.Atoi(voteStr)
			require.NoErrorf(tb, err, "failed to parse ballot: %w", err)
			cdt := candidatesByIndex[candidateIndex]
			require.NotNilf(tb, cdt, "malformed ballot %q: invalid candidate index %d", line, candidateIndex)
			ballot.RankedCandidates = append(ballot.RankedCandidates, cdt)
		}

		for i := 0; i < weight; i++ {
			ballots = append(ballots, ballot)
		}
	}
	require.NoError(tb, sc.Err())

	for i := 0; i < numCandidates; i++ {
		require.True(tb, sc.Scan())
		candidatesByIndex[i+1].Name = sc.Text()
	}

	require.NoError(tb, sc.Err())

	for _, cdt := range candidatesByIndex {
		candidates = append(candidates, cdt)
	}

	return candidates, ballots, numSeats
}

func AssertVoteWinners(tb testing.TB, results *ElectionResults, expected []string) {
	winners := results.GetWinners()
	assert.Equal(tb, len(expected), len(winners), "non-equal winner count")
	for i, winner := range winners {
		assert.Truef(tb, slices.Contains(expected, winner.Name), "winner %d (%q) not in expected list", i, winner.Name)
	}
	for i, expectedWinner := range expected {
		assert.Truef(tb, slices.ContainsFunc(winners, func(c *Candidate) bool {
			return c.Name == expectedWinner
		}), "winner %d (%q) not in actual list", i, expectedWinner)
	}
}

func TestSTVScotland2022(t *testing.T) {
	candidates, ballots, numSeats := ParseBLT(t, "testdata/Scotland2022_Ward_1_Penicuik.blt")

	results, err := SingleTransferableVote(candidates, ballots, numSeats, DefaultSingleTransferableVoteOptions())
	if err != nil {
		t.Fatal(err)
	}
	AssertVoteWinners(t, results, []string{
		`"Debbi MCCALL" "Scottish National Party (SNP)"`,
		`"Willie MCEWAN" "Scottish Labour Party"`,
		`"Connor MCMANUS" "Scottish National Party (SNP)"`,
	})
}