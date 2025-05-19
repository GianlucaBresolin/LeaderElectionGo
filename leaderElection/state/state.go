package state

type State struct {
	value       string
	CandidateCh chan struct{}
	LeaderCh    chan struct{}
	FollowerCh  chan struct{}
}

func NewState() *State {
	candidateCh := make(chan struct{})
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
				state.value = "candidate"
				state.handleCandidateRequest()
			case <-leaderCh:
				state.value = "leader"
				state.handleLeaderRequest()
			case <-followerCh:
				state.value = "follower"
				state.handleFollowerRequest()
			}
		}
	}()

	return &state
}
