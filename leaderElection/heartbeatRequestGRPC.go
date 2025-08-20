package leaderElection

import (
	"context"
	"log"

	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/myVote"
	pb "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
)

func (node *Node) HeartbeatRequestGRPC(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	// get the current term
	termResponseCh := make(chan int)
	node.currentTerm.GetTermReq <- term.GetTermSignal{
		ResponseCh: termResponseCh,
	}
	currentTerm := <-termResponseCh

	heartbeatResponse := &pb.HeartbeatResponse{
		Term:    int32(currentTerm),
		Success: false, // default
	}

	switch {
	case int(req.Term) < currentTerm:
		// stale request, do not grant success (default)
		return heartbeatResponse, nil
	case int(req.Term) == currentTerm:
		// current term matches, proceed to check the leader
	case int(req.Term) > currentTerm:
		// new term, update the current term
		node.currentTerm.SetTermReq <- term.SetTermSignal{
			Value: int(req.Term),
		}
		// revert to follower state
		node.state.FollowerCh <- state.FollowerSignal{
			ElectionTimerRef: node.electionTimer,
			StopLeadershipCh: node.stopLeadershipCh,
			Term:             int(req.Term),
		}
		// reset my vote
		node.myVote.ResetReq <- myVote.ResetSignal{
			Term: int(req.Term),
		}
		// proceed to set the leader
	}

	// try to set the currentLeader
	setCurrentLeaderResponseCh := make(chan string)
	node.currentLeader.SetCurrentLeaderReq <- currentLeader.SetCurrentLeaderSignal{
		Leader:     req.Leader,
		ResponseCh: setCurrentLeaderResponseCh,
	}
	currentLeader := <-setCurrentLeaderResponseCh

	if currentLeader == req.Leader {
		// the current leader is set successfully or was already set to this leader
		// grant the heartbeat's success
		heartbeatResponse.Success = true

		// reset the election timer
		node.electionTimer.ResetReq <- electionTimer.ResetSignal{}

		log.Println("Node", node.ID, ": Heartbeat received from leader:", req.Leader, "for term:", req.Term)
	}
	// else not the current leader, do not grant success (default)
	return heartbeatResponse, nil
}
