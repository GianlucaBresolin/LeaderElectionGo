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
	candidateCh := make(chan CandidateSignal)
	leaderCh := make(chan LeaderSignal)
	followerCh := make(chan FollowerSignal)

	state := State{
		value:       "follower", // initally the state is set to be follwer
		FollowerCh:  followerCh,
		CandidateCh: candidateCh,
		LeaderCh:    leaderCh,
	}

	go func() {
		for {
			select {
			case <-followerCh:
				state.value = "follower"
			case <-candidateCh:
				state.value = "candidate"
			case <-leaderCh:
				state.value = "leader"
			}
		}
	}()

	return &state
}
