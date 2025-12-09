package voting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type Candidate struct {
	Name string
}

func NewCandidate(name string) *Candidate {
	return &Candidate{name}
}

func (c *Candidate) String() string {
	return fmt.Sprintf("Candidate('%s')", c.Name)
}

func (c *Candidate) Hash() string {
	hasher := sha256.New()
	hasher.Write([]byte(c.Name))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (c *Candidate) Equals(a *Candidate) bool {
	return a.Name == c.Name
}

type Ballot struct {
	RankedCandidates []*Candidate
}

func NewBallot(candidates []*Candidate) *Ballot {
	if hasDuplicates(candidates) {
		panic("duplicate candidate")
	}
	return &Ballot{candidates}
}

func hasDuplicates(candidates []*Candidate) bool {
	visited := make(map[string]bool)
	for _, c := range candidates {
		_, ok := visited[c.Name]
		if ok {
			return true
		}
		visited[c.Name] = true
	}
	return false
}

func (b *Ballot) String() string {
	candidateNames := make([]string, 0, len(b.RankedCandidates))
	for _, c := range b.RankedCandidates {
		candidateNames = append(candidateNames, c.Name)
	}
	candidates := strings.Join(candidateNames, ", ")
	return fmt.Sprintf("Ballot(%s)", candidates)
}
