package state

type CandidateSignal struct{}

type State struct {
	value       string
	CandidateCh chan CandidateSignal
	LeaderCh    chan struct{}
	FollowerCh  chan struct{}
}

func NewState() *State {
	candidateCh := make(chan CandidateSignal)
	leaderCh := make(chan struct{})
	followerCh := make(chan struct{})

	state := State{
		value:       "follower", // initally the state is set to be follwer
		CandidateCh: candidateCh,
		LeaderCh:    leaderCh,
		FollowerCh:  followerCh,
	}

	go func() {
		for {
			select {
			case <-candidateCh:
				state.handleCandidateRequest()
			case <-leaderCh:
				state.handleLeaderRequest()
			case <-followerCh:
				state.handleFollowerRequest()
			}
		}
	}()

	return &state
}
