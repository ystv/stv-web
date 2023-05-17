package voting

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

//go:embed testdata/*.blt
var testdata embed.FS

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
