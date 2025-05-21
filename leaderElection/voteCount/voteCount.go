package voteCount

import "log"

type AddVoteArg struct {
	term    int
	voterID string
}
type ResetArg struct {
	term int
}

type VoteCount struct {
	voteCount  int
	term       int
	voterMap   map[string]struct{}
	AddVoteReq chan AddVoteArg
	ResetReq   chan ResetArg
}

func NewVoteCount() *VoteCount {
	addVoteReq := make(chan AddVoteArg)
	resetReq := make(chan ResetArg)

	voteCount := VoteCount{
		voteCount:  0,
		term:       0,
		voterMap:   make(map[string]struct{}),
		AddVoteReq: addVoteReq,
		ResetReq:   resetReq,
	}

	go func() {
		for {
			select {
			case arg := <-addVoteReq:
				voteCount.addVote(arg)
			case arg := <-resetReq:
				voteCount.reset(arg)
			}
		}
	}()

	return &voteCount
}

func (voteCount *VoteCount) addVote(arg AddVoteArg) {
	if arg.term != voteCount.term {
		log.Println("Vote count add request ignored: term is not equal to current term")
	} else {
		if _, exists := voteCount.voterMap[arg.voterID]; exists {
			// idempotency behavior: if the voter has already voted, do nothing
			return
		}
		voteCount.voteCount++
		voteCount.voterMap[arg.voterID] = struct{}{}
	}
}

func (voteCount *VoteCount) reset(arg ResetArg) {
	if arg.term > voteCount.term {
		voteCount.term = arg.term
		voteCount.voteCount = 0
		voteCount.voterMap = make(map[string]struct{})
	} else {
		log.Println("Vote count reset request ignored: term is not greater than current term")
	}
}
