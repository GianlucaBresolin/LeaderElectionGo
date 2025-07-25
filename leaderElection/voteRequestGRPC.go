package leaderElection

import (
	"context"
	"log"

	"LeaderElectionGo/leaderElection/myVote"
	pb "LeaderElectionGo/leaderElection/services/voteRequest/voteRequestService"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/utils"
)

func (node *Node) VoteRequestGRPC(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	// response
	voteResponse := &pb.VoteResponse{
		Term: req.Term,
		// the vote is not granted by default
		Granted: false,
	}

	// read the current term
	currentTermCh := make(chan int)
	node.currentTerm.GetTermReq <- term.GetTermSignal{
		ResponseCh: currentTermCh,
	}
	currentTerm := <-currentTermCh

	// check if the term is valid
	if int(req.Term) < currentTerm {
		// the term is not valid, not grant the vote
		return voteResponse, nil
	}

	// if the term is greater, update the current term and revert to follower state
	if int(req.Term) > currentTerm {
		node.currentTerm.SetTermReq <- term.SetTermSignal{
			Value: int(req.Term),
		}

		// revert to follower state
		node.state.FollowerCh <- state.FollowerSignal{
			HeartbeatTimerRef: node.heartbeatTimer,
			ElectionTimerRef:  node.electionTimer,
			StopLeadershipCh:  node.stopLeadershipCh,
			Term:              int(req.Term),
		}
	}

	//the term is valid, try to set the our vote
	responseCh := make(chan bool)
	node.myVote.SetVoteReq <- myVote.SetVoteSignal{
		Vote:       utils.NodeID(req.CandidateId),
		Term:       int(req.Term),
		ResponseCh: responseCh,
	}
	if success := <-responseCh; success {
		// we set the vote successfully
		voteResponse.Granted = true
	}
	// else we do nothing (vote not granted)
	log.Println("NODE", node.ID, "VOTE REQUEST FROM NODE", req.CandidateId, "FOR TERM", req.Term, "GRANTED:", voteResponse.Granted)
	return voteResponse, nil
}
