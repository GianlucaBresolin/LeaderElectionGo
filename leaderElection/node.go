package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/voteCount"

	pb "LeaderElectionGo/leaderElection/voteRequestService"
)

type Node struct {
	ID            string
	address       string
	state         *state.State
	currentTerm   *term.Term
	electionTimer *electionTimer.ElectionTimer
	voteCount     *voteCount.VoteCount
	myVote        *myVote.MyVote

	configurationMap map[string]string // map[nodeID]address

	pb.UnimplementedVoteRequestServiceServer
}

func NewNode(id string, address string, configurationMap map[string]string) *Node {
	node := Node{
		ID:               id,
		address:          address,
		state:            state.NewState(),
		currentTerm:      term.NewTerm(),
		electionTimer:    electionTimer.NewElectionTimer(150, 300),
		voteCount:        voteCount.NewVoteCount(),
		myVote:           myVote.NewMyVote(),
		configurationMap: configurationMap,
	}

	//node.prepareConnections()

	go node.run()

	return &node
}
