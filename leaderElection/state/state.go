package state

type GetStateResponse struct {
	State string
	Term  int
}
type GetStateSignal struct {
	ResponseCh chan<- GetStateResponse
}
type FollowerSignal struct {
	Term       int
	ResponseCh chan<- bool
}
type CandidateSignal struct {
	Term       int
	ResponseCh chan<- bool
}
type LeaderSignal struct {
	Term       int
	ResponseCh chan<- bool
}

type State struct {
	term        int
	value       string
	GetStateCh  chan GetStateSignal
	FollowerCh  chan FollowerSignal
	CandidateCh chan CandidateSignal
	LeaderCh    chan LeaderSignal
}

func NewState() *State {
	state := &State{
		value:       "follower", // initally the state is set to be follwer
		GetStateCh:  make(chan GetStateSignal),
		FollowerCh:  make(chan FollowerSignal),
		CandidateCh: make(chan CandidateSignal),
		LeaderCh:    make(chan LeaderSignal),
	}

	go func() {
		for {
			select {
			case signal := <-state.FollowerCh:
				state.setFollower(signal)
			case signal := <-state.CandidateCh:
				state.setCandidate(signal)
			case signal := <-state.LeaderCh:
				state.setLeader(signal)
			case signal := <-state.GetStateCh:
				state.GetState(signal)
			}
		}
	}()

	return state
}

func (s *State) GetState(signal GetStateSignal) {
	signal.ResponseCh <- GetStateResponse{
		State: s.value,
		Term:  s.term,
	}
}

func (s *State) setFollower(signal FollowerSignal) {
	switch {
	case signal.Term > s.term:
		s.value = "follower"
		s.term = signal.Term
		signal.ResponseCh <- true
	case signal.Term == s.term:
		if s.value != "follower" {
			s.value = "follower"
			signal.ResponseCh <- true
		} else {
			signal.ResponseCh <- false
		}
	default:
		// else stale request, ignore it
		signal.ResponseCh <- false
	}
}

func (s *State) setCandidate(signal CandidateSignal) {
	switch {
	case s.term < signal.Term:
		s.term = signal.Term
		s.value = "candidate"
		signal.ResponseCh <- true
	case s.term == signal.Term:
		if s.value == "follower" {
			s.value = "candidate"
			signal.ResponseCh <- true
		}
		// else already a candidate or a leader for this term
		signal.ResponseCh <- false
	default:
		// stale request
		signal.ResponseCh <- false
	}
}

func (s *State) setLeader(signal LeaderSignal) {
	switch {
	case s.term < signal.Term:
		s.term = signal.Term
		s.value = "leader"
		signal.ResponseCh <- true
		return
	case s.term == signal.Term:
		if s.value != "leader" {
			s.value = "leader"
			signal.ResponseCh <- true
			return
		} else {
			signal.ResponseCh <- false
			return
		}
	default:
		// stale request
		signal.ResponseCh <- false
	}
}
