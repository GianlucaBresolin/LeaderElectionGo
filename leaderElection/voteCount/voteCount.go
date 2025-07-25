package voteCount

import (
	"LeaderElectionGo/leaderElection/utils"
	"log"
)

type AddVoteSignal struct {
	Term           int
	VoterID        utils.NodeID
	BecomeLeaderCh chan<- bool
}
type ResetSignal struct {
	Term int
}

type VoteCount struct {
	voteCount  int
	term       int
	voterMap   map[utils.NodeID]bool
	leaderFlag bool
	AddVoteReq chan AddVoteSignal
	ResetReq   chan ResetSignal
}

func NewVoteCount(configurationList []utils.NodeID) *VoteCount {
	voteCount := &VoteCount{
		voteCount:  0,
		term:       0,
		voterMap:   make(map[utils.NodeID]bool),
		leaderFlag: false,
		AddVoteReq: make(chan AddVoteSignal),
		ResetReq:   make(chan ResetSignal),
	}

	// build the voterMap from teh configurationMap
	for _, voterID := range configurationList {
		voteCount.voterMap[voterID] = false
	}

	go func() {
		for {
			select {
			case signal := <-voteCount.AddVoteReq:
				voteCount.addVote(signal)
			case signal := <-voteCount.ResetReq:
				voteCount.reset(signal)
			}
		}
	}()

	return voteCount
}

func (voteCount *VoteCount) addVote(signal AddVoteSignal) {
	switch {
	case signal.Term < voteCount.term:
		// stale request, do nothing
		return
	case signal.Term > voteCount.term:
		// new term, proceed to reset the vote count
		voteCount.reset(ResetSignal{Term: signal.Term})
	default:
		// signal.Term == voteCount.term: proceed to add the vote
	}
	// now the signal.Term is guaranteed to be == voteCount.term
	if votedFlag := voteCount.voterMap[signal.VoterID]; votedFlag {
		// idempotency behavior: if the voter has already voted, do nothing
		return
	}
	voteCount.voteCount++
	voteCount.voterMap[signal.VoterID] = true

	if voteCount.voteCount > len(voteCount.voterMap)/2 && !voteCount.leaderFlag {
		log.Printf("Reached majority of votes in the cluster.")
		voteCount.leaderFlag = true
		signal.BecomeLeaderCh <- true
		return
	}
	// not enough votes to become leader yet or already a leader
	signal.BecomeLeaderCh <- false
}

func (voteCount *VoteCount) reset(signal ResetSignal) {
	if signal.Term > voteCount.term {
		voteCount.term = signal.Term
		voteCount.voteCount = 0
		for voterID := range voteCount.voterMap {
			voteCount.voterMap[voterID] = false
		}
		voteCount.leaderFlag = false
	}
	// else the reset for that term already happened, do nothing
}
