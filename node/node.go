package node

import (
	pb "LeaderElectionGo/node/voteRequestService"
)

type Node struct {
	ID      string
	address string
	pb.UnimplementedVoteRequestServiceServer
}

func NewNode(id string, address string) *Node {
	return &Node{
		ID:      id,
		address: address,
	}
}
