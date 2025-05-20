package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"

	pb "LeaderElectionGo/leaderElection/voteRequestService"
)

type Node struct {
	ID            string
	address       string
	state         *state.State
	currentTerm   *term.Term
	electionTimer *electionTimer.ElectionTimer

	pb.UnimplementedVoteRequestServiceServer
}

func NewNode(id string, address string) *Node {
	return &Node{
		ID:            id,
		address:       address,
		state:         state.NewState(),
		currentTerm:   term.NewTerm(),
		electionTimer: electionTimer.NewElectionTimer(150, 300),
	}
}
