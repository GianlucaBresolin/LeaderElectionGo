package myVote

type SetVoteSignal struct {
	Vote       string
	ResponseCh chan<- bool
}
type ResetSignal struct{}
type ReadMyVoteSignal struct {
	ResponseCh chan<- string
}

type MyVote struct {
	myVote        string
	SetVoteReq    chan SetVoteSignal
	ResetReq      chan ResetSignal
	ReadMyVoteReq chan ReadMyVoteSignal
}

func NewMyVote() *MyVote {
	setVoteReq := make(chan SetVoteSignal)

	myVote := MyVote{
		myVote:     "",
		SetVoteReq: setVoteReq,
	}

	go func() {
		for {
			select {
			case signal := <-setVoteReq:
				myVote.setVote(signal)
			case <-myVote.ResetReq:
				myVote.myVote = ""
			case signal := <-myVote.ReadMyVoteReq:
				signal.ResponseCh <- myVote.myVote
			}
		}
	}()

	return &myVote
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
