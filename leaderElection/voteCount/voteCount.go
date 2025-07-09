package voteCount

import "log"

type AddVoteSignal struct {
	Term           int
	VoterID        string
	BecomeLeaderCh chan bool
}
type ResetSignal struct {
	Term int
}

type VoteCount struct {
	voteCount  int
	term       int
	voterMap   map[string]bool
	leaderFlag bool
	AddVoteReq chan AddVoteSignal
	ResetReq   chan ResetSignal
}

func NewVoteCount(configurationMap map[string]string) *VoteCount {
	voteCount := &VoteCount{
		voteCount:  0,
		term:       0,
		voterMap:   make(map[string]bool),
		leaderFlag: false,
		AddVoteReq: make(chan AddVoteSignal),
		ResetReq:   make(chan ResetSignal),
	}

	// build the voterMap from teh configurationMap
	for voterID := range configurationMap {
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
	if signal.Term != voteCount.term {
		log.Println("Vote count add request ignored: term is not equal to current term")
	} else {
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
}

func (voteCount *VoteCount) reset(signal ResetSignal) {
	if signal.Term > voteCount.term {
		voteCount.term = signal.Term
		voteCount.voteCount = 0
		for voterID := range voteCount.voterMap {
			voteCount.voterMap[voterID] = false
		}
		voteCount.leaderFlag = false
	} else {
		log.Println("Vote count reset request ignored: term is not greater than current term")
	}
}
