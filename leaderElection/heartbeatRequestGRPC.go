package leaderElection

import (
	"context"
	"log"

	"LeaderElectionGo/leaderElection/allowVotes"
	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
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
		// new term, try to update the current term
		successSetTermCh := make(chan bool)
		node.currentTerm.SetTermReq <- term.SetTermSignal{
			Value:      int(req.Term),
			ResponseCh: successSetTermCh,
		}
		if success := <-successSetTermCh; !success {
			// we failed to set the term, i.e., the request has become stale
			// do not grant success (default)
			return heartbeatResponse, nil
		}
		// revert to follower state
		successRevertToFollowerCh := make(chan bool)
		node.state.FollowerCh <- state.FollowerSignal{
			Term:       int(req.Term),
			ResponseCh: successRevertToFollowerCh,
		}
		if success := <-successRevertToFollowerCh; success {
			// reset the election timer
			node.electionTimer.ResetReq <- electionTimer.ResetSignal{
				Term: int(req.Term),
			}
		}
	}

	// try to set the currentLeader
	setCurrentLeaderResponseCh := make(chan currentLeader.CurrentLeaderResponse)
	node.currentLeader.SetCurrentLeaderReq <- currentLeader.SetCurrentLeaderSignal{
		Leader:     req.Leader,
		Term:       int(req.Term),
		ResponseCh: setCurrentLeaderResponseCh,
	}
	response := <-setCurrentLeaderResponseCh

	if response.Leader == req.Leader && response.Term == int(req.Term) {
		heartbeatResponse.Success = true
		// disallow votes
		node.allowVotes.DisallowCh <- allowVotes.DisallowSignal{}
		// reset the election timer
		node.electionTimer.ResetReq <- electionTimer.ResetSignal{
			Term: int(req.Term),
		}
		log.Println("Node", node.ID, ": Heartbeat received from leader:", req.Leader, "for term:", req.Term)
	}
	// else not the current leader, do not grant success (default)
	return heartbeatResponse, nil
}
