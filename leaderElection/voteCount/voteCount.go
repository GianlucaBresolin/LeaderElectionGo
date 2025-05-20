package voteCount

type AddVoteSignal struct{}
type ResetSignal struct{}

type VoteCount struct {
	voteCount  int
	AddVoteReq chan AddVoteSignal
	ResetReq   chan ResetSignal
}

func NewVoteCount() *VoteCount {
	addVoteReq := make(chan AddVoteSignal)
	resetReq := make(chan ResetSignal)

	voteCount := VoteCount{
		voteCount:  0,
		AddVoteReq: addVoteReq,
		ResetReq:   resetReq,
	}

	go func() {
		for {
			select {
			case <-addVoteReq:
				voteCount.addVote()
			case <-resetReq:
				voteCount.reset()
			}
		}
	}()

	return &voteCount
}

func (voteCount *VoteCount) addVote() {
	voteCount.voteCount++
}

func (voteCount *VoteCount) reset() {
	voteCount.voteCount = 0
}
