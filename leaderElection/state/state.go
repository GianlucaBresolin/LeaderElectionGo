package state

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/internalUtils"
	"log"
)

type FollowerSignal struct {
	ElectionTimerRef *electionTimer.ElectionTimer
	StopLeadershipCh chan<- internalUtils.StopLeadershipSignal
	Term             int
}
type CandidateSignal struct {
	Term int
}
type LeaderSignal struct {
	Term       int
	ResponseCh chan<- bool
}

type State struct {
	term        int
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
			case signal := <-state.FollowerCh:
				state.setFollower(signal)
			case signal := <-state.CandidateCh:
				state.setCandidate(signal)
			case signal := <-state.LeaderCh:
				state.setLeader(signal)
			}
		}
	}()

	return state
}

func (s *State) setFollower(signal FollowerSignal) {
	if signal.Term >= s.term {
		if s.value == "leader" {
			if signal.StopLeadershipCh == nil || signal.ElectionTimerRef == nil {
				log.Fatal("Error: StopLeadershipCh, HeartbeatTimerRef, or ElectionTimerRef is nil, cannot revert to folloer. (Incongruent state)")
			}
			// restart the election timer and
			signal.ElectionTimerRef.ResetReq <- electionTimer.ResetSignal{}
			// stop the leadership
			signal.StopLeadershipCh <- internalUtils.StopLeadershipSignal{}
		}
		s.value = "follower"
		s.term = signal.Term
	}
	// else stale request, ignore it
}

func (s *State) setCandidate(signal CandidateSignal) {
	switch {
	case s.term < signal.Term:
		s.term = signal.Term
		s.value = "candidate"
	case s.term == signal.Term:
		if s.value == "follower" {
			s.value = "candidate"
		}
		// else do nothing, already a candidate or a leader for this term
	default:
		// stale request, do nothing
	}
}

func (s *State) setLeader(signal LeaderSignal) {
	switch {
	case s.term < signal.Term:
		s.term = signal.Term
		s.value = "leader"
		signal.ResponseCh <- true // send true if the term is valid
		return
	case s.term == signal.Term:
		if s.value != "leader" {
			s.value = "leader"
			signal.ResponseCh <- true // send true if the term is valid
			return
		} else {
			signal.ResponseCh <- false // already a leader for this term
			return
		}
	default:
		signal.ResponseCh <- false // send false if the term is not valid
		return
	}
}
