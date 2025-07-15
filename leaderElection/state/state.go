package state

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/heartbeatTimer"
	"LeaderElectionGo/leaderElection/stopLeadershipSignal"
	"log"
)

type FollowerSignal struct {
	HeartbeatTimerRef *heartbeatTimer.HeartbeatTimer
	ElectionTimerRef  *electionTimer.ElectionTimer
	StopLeadershipCh  chan stopLeadershipSignal.StopLeadershipSignal
}
type CandidateSignal struct{}
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
			case <-state.CandidateCh:
				state.value = "candidate"
			case signal := <-state.LeaderCh:
				state.setLeader(signal)
			}
		}
	}()

	return state
}

func (s *State) setFollower(signal FollowerSignal) {
	if s.value == "leader" {
		if signal.StopLeadershipCh == nil || signal.HeartbeatTimerRef == nil || signal.ElectionTimerRef == nil {
			log.Fatal("Error: StopLeadershipCh, HeartbeatTimerRef, or ElectionTimerRef is nil, cannot revert to folloer. (Incongruent state)")
		}
		// if the current state is leader, stop the heartbeat timer
		signal.HeartbeatTimerRef.StopReq <- heartbeatTimer.StopSignal{}
		// restart the election timer
		signal.ElectionTimerRef.ResetReq <- electionTimer.ResetSignal{}
		// stop the leadership
		signal.StopLeadershipCh <- stopLeadershipSignal.StopLeadershipSignal{}
	}
	s.value = "follower"
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
