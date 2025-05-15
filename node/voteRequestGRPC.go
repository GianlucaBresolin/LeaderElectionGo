package node

import (
	"context"

	pb "LeaderElectionGo/node/voteRequestService"
)

func (self *Node) Vote(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	return &pb.VoteResponse{
		Granted: true,
	}, nil
}
