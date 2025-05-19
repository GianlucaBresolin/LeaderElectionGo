package leaderElection

import (
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"

	pb "LeaderElectionGo/leaderElection/voteRequestService"
)

type Node struct {
	ID          string
	address     string
	currentTerm *term.Term
	state       *state.State
	pb.UnimplementedVoteRequestServiceServer
}

func NewNode(id string, address string) *Node {
	return &Node{
		ID:          id,
		address:     address,
		currentTerm: term.NewTerm(),
		state:       state.NewState(),
	}
}
