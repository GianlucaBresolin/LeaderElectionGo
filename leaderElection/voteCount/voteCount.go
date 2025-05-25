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
	voterMap   map[string]bool
	AddVoteReq chan AddVoteSignal
	ResetReq   chan ResetSignal
}

func NewVoteCount(configurationMap map[string]string) *VoteCount {
	addVoteReq := make(chan AddVoteSignal)
	resetReq := make(chan ResetSignal)

	voteCount := &VoteCount{
		voteCount:  0,
		term:       0,
		voterMap:   make(map[string]bool),
		AddVoteReq: addVoteReq,
		ResetReq:   resetReq,
	}

	// build the voterMap from teh configurationMap
	for voterID := range configurationMap {
		voteCount.voterMap[voterID] = false
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
		if votedFlag := voteCount.voterMap[signal.VoterID]; votedFlag {
			// idempotency behavior: if the voter has already voted, do nothing
			return
		}
		voteCount.voteCount++
		voteCount.voterMap[signal.VoterID] = true

		if voteCount.voteCount > len(voteCount.voterMap)/2 {
			log.Printf("Reached majority of votes in the cluster.")
		}
	}
}

func (voteCount *VoteCount) reset(signal ResetSignal) {
	if signal.Term > voteCount.term {
		voteCount.term = signal.Term
		voteCount.voteCount = 0
		for voterID := range voteCount.voterMap {
			voteCount.voterMap[voterID] = false
		}
	} else {
		log.Println("Vote count reset request ignored: term is not greater than current term")
	}
}
