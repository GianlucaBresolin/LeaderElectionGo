package myVote

type SetVoteSignal struct {
	Vote       string
	ResponseCh chan<- bool
}
type ResetSignal struct{}

type MyVote struct {
	myVote     string
	SetVoteReq chan SetVoteSignal
	ResetReq   chan ResetSignal
}

func NewMyVote() *MyVote {
	myVote := &MyVote{
		myVote:     "",
		SetVoteReq: make(chan SetVoteSignal),
		ResetReq:   make(chan ResetSignal),
	}

	go func() {
		for {
			select {
			case signal := <-myVote.SetVoteReq:
				myVote.setVote(signal)
			case <-myVote.ResetReq:
				myVote.myVote = ""
			}
		}
	}()

	return myVote
}

func (myVote *MyVote) setVote(signal SetVoteSignal) {
	if myVote.myVote != "" {
		// we altready voted
		signal.ResponseCh <- false
		return
	}
	// set the vote successfully
	myVote.myVote = signal.Vote
	signal.ResponseCh <- true
}
