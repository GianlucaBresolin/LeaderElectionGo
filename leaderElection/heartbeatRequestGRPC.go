package leaderElection

import (
	"context"

	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
	pb "LeaderElectionGo/leaderElection/heartbeatRequestService"
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

	if req.Term >= heartbeatResponse.Term {
		// try to set the currentLeader
		setCurrentLeaderResponseCh := make(chan string)
		node.currentLeader.SetCurrentLeaderReq <- currentLeader.SetCurrentLeaderSignal{
			Leader:     req.Leader,
			ResponseCh: setCurrentLeaderResponseCh,
		}
		currentLeader := <-setCurrentLeaderResponseCh
		if currentLeader == req.Leader {
			// the current leader is set successfully or was altresady set to this leader
			// grant the heartbeat's success
			heartbeatResponse.Success = true

			// reset the election timer
			node.electionTimer.ResetReq <- electionTimer.ResetSignal{}
		}
		// else not the current leader, do not grant success (default)
	}
	// else req.Term <  currentTerm, stale request -> do not grant success (default)
	return heartbeatResponse, nil
}
