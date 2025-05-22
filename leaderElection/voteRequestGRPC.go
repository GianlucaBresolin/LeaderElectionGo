package leaderElection

import (
	"context"

	"LeaderElectionGo/leaderElection/myVote"
	pb "LeaderElectionGo/leaderElection/voteRequestService"
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
	node.currentTerm.GetTermReq <- currentTermCh
	currentTerm := <-currentTermCh

	// check if the term is valid
	if int(req.Term) < currentTerm {
		// the term is not valid, not grant the vote
		return voteResponse, nil
	}

	//the term is valid, try to set the our vote
	responseCh := make(chan bool)
	node.myVote.SetVoteReq <- myVote.SetVoteSignal{
		Vote:       req.CandidateId,
		ResponseCh: responseCh,
	}
	if success := <-responseCh; success {
		// we set the vote successfully
		voteResponse.Granted = true
	}
	// else we do nothing (vote not granted)
	return voteResponse, nil
}
