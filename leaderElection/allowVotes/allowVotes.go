package allowVotes

import (
	"time"
)

const MIN_ELECTION_TIMEOUT = 150 * time.Millisecond

type RestartSignal struct{}
type StopSignal struct{}
type DisallowSignal struct{}
type GetAllowVotesCh struct {
	ResponseCh chan bool
}

type AllowVotes struct {
	value           bool
	StopCh          chan StopSignal
	RestartCh       chan RestartSignal
	DisallowCh      chan DisallowSignal
	GetAllowVotesCh chan GetAllowVotesCh
}

func NewAllowVotes() *AllowVotes {
	allowVotes := &AllowVotes{
		value:           true, // initially votes are allowed
		StopCh:          make(chan StopSignal),
		RestartCh:       make(chan RestartSignal),
		DisallowCh:      make(chan DisallowSignal),
		GetAllowVotesCh: make(chan GetAllowVotesCh),
	}

	go func() {
		for {
			select {
			case <-allowVotes.StopCh:
				<-allowVotes.RestartCh
			case <-allowVotes.DisallowCh:
				allowVotes.value = false
			case <-time.After(MIN_ELECTION_TIMEOUT):
				allowVotes.value = true
			case signal := <-allowVotes.GetAllowVotesCh:
				signal.ResponseCh <- allowVotes.value
			}
		}
	}()

	return allowVotes
}
