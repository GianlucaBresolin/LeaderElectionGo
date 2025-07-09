package myVote

type SetVoteSignal struct {
	Vote       string
	Term       int
	ResponseCh chan<- bool
}
type ResetSignal struct {
	Term int
}

type MyVote struct {
	term       int
	myVote     string
	SetVoteReq chan SetVoteSignal
	ResetReq   chan ResetSignal
}

func NewMyVote() *MyVote {
	myVote := &MyVote{
		myVote:     "",
		term:       0,
		SetVoteReq: make(chan SetVoteSignal),
		ResetReq:   make(chan ResetSignal),
	}

	go func() {
		for {
			select {
			case signal := <-myVote.SetVoteReq:
				myVote.setVote(signal)
			case signal := <-myVote.ResetReq:
				myVote.reset(signal)
			}
		}
	}()

	return myVote
}

func (myVote *MyVote) setVote(signal SetVoteSignal) {
	if signal.Term < myVote.term || (signal.Term == myVote.term && myVote.myVote != "") {
		// the term is not valid or we already voted in this term
		signal.ResponseCh <- false
		return
	}
	// set the vote successfully
	myVote.term = signal.Term
	myVote.myVote = signal.Vote
	signal.ResponseCh <- true
}

func (myVote *MyVote) reset(signal ResetSignal) {
	// reset only if the provided term is greater than the current term
	if myVote.term < signal.Term {
		myVote.myVote = ""
		myVote.term = signal.Term
	}
	// else do nothing, we are already in a higher term
}
