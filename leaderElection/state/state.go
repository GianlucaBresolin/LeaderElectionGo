package state

type FollowerSignal struct{}
type CandidateSignal struct{}
type LeaderSignal struct{}

type State struct {
	value       string
	FollowerCh  chan FollowerSignal
	CandidateCh chan CandidateSignal
	LeaderCh    chan LeaderSignal
}

func NewState() *State {
	state := &State{
		value:       "follower", // initally the state is set to be follwer
		FollowerCh:  make(chan FollowerSignal),
		CandidateCh: make(chan CandidateSignal),
		LeaderCh:    make(chan LeaderSignal),
	}

	go func() {
		for {
			select {
			case <-state.FollowerCh:
				state.value = "follower"
			case <-state.CandidateCh:
				state.value = "candidate"
			case <-state.LeaderCh:
				state.value = "leader"
			}
		}
	}()

	return state
}
