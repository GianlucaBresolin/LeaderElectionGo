package myVote

import "LeaderElectionGo/leaderElection/utils"

type SetVoteSignal struct {
	Vote       utils.NodeID
	Term       int
	ResponseCh chan<- bool
}

type MyVote struct {
	term       int
	myVote     utils.NodeID
	SetVoteReq chan SetVoteSignal
}

func NewMyVote() *MyVote {
	myVote := &MyVote{
		myVote:     "",
		term:       0,
		SetVoteReq: make(chan SetVoteSignal),
	}

	go func() {
		for {
			select {
			case signal := <-myVote.SetVoteReq:
				myVote.setVote(signal)
			}
		}
	}()

	return myVote
}

func (myVote *MyVote) setVote(signal SetVoteSignal) {
	if signal.Term < myVote.term || (signal.Term == myVote.term && !(myVote.myVote == "" || myVote.myVote == signal.Vote)) {
		// the term is not valid or we already voted in this term for someone else
		signal.ResponseCh <- false
		return
	}
	// set the vote successfully (and possibly update the term)
	myVote.term = signal.Term
	myVote.myVote = signal.Vote
	signal.ResponseCh <- true
}
