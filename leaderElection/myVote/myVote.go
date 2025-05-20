package myVote

type SetVoteArg struct {
	ID string
}

type MyVote struct {
	myVote     string
	SetVoteReq chan SetVoteArg
}

func NewMyVote() *MyVote {
	setVoteReq := make(chan SetVoteArg)

	myVote := MyVote{
		myVote:     "",
		SetVoteReq: setVoteReq,
	}

	go func() {
		for {
			select {
			case arg := <-setVoteReq:
				myVote.setVote(arg)
			}
		}
	}()

	return &myVote
}

func (myVote *MyVote) setVote(arg SetVoteArg) {
	myVote.myVote = arg.ID
}
