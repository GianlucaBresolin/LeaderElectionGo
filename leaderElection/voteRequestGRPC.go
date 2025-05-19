package leaderElection

import (
	"context"

	pb "LeaderElectionGo/leaderElection/voteRequestService"
)

func (self *Node) Vote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {

	// TODO

	return &pb.VoteResponse{
		Granted: true,
	}, nil
}
