package voteCount

import "log"

type AddVoteSignal struct {
	Term    int
	VoterID string
}
type ResetSignal struct {
	Term int
}

type VoteCount struct {
	voteCount  int
	term       int
	voterMap   map[string]struct{}
	AddVoteReq chan AddVoteSignal
	ResetReq   chan ResetSignal
}

func NewVoteCount() *VoteCount {
	addVoteReq := make(chan AddVoteSignal)
	resetReq := make(chan ResetSignal)

	voteCount := &VoteCount{
		voteCount:  0,
		term:       0,
		voterMap:   make(map[string]struct{}),
		AddVoteReq: addVoteReq,
		ResetReq:   resetReq,
	}

	go func() {
		for {
			select {
			case signal := <-addVoteReq:
				voteCount.addVote(signal)
			case signal := <-resetReq:
				voteCount.reset(signal)
			}
		}
	}()

	return voteCount
}

func (voteCount *VoteCount) addVote(signal AddVoteSignal) {
	if signal.Term != voteCount.term {
		log.Println("Vote count add request ignored: term is not equal to current term")
	} else {
		if _, exists := voteCount.voterMap[signal.VoterID]; exists {
			// idempotency behavior: if the voter has already voted, do nothing
			return
		}
		voteCount.voteCount++
		voteCount.voterMap[signal.VoterID] = struct{}{}
	}
}

func (voteCount *VoteCount) reset(signal ResetSignal) {
	if signal.Term > voteCount.term {
		voteCount.term = signal.Term
		voteCount.voteCount = 0
		voteCount.voterMap = make(map[string]struct{})
	} else {
		log.Println("Vote count reset request ignored: term is not greater than current term")
	}
}
