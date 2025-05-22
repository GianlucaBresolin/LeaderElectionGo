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

	configurationMap map[string]connectionData // map[nodeID]connectionData
	CloseCh          chan CloseSignal

	pb.UnimplementedVoteRequestServiceServer
}

func NewNode(id string, address string, configurationMap map[string]connectionData) *Node {
	node := Node{
		ID:               id,
		address:          address,
		state:            state.NewState(),
		currentTerm:      term.NewTerm(),
		electionTimer:    electionTimer.NewElectionTimer(150, 300),
		voteCount:        voteCount.NewVoteCount(),
		myVote:           myVote.NewMyVote(),
		configurationMap: configurationMap,
		CloseCh:          make(chan CloseSignal),
	}

	node.prepareServer()
	node.prepareConnections()

	go node.run()

	return &node
}
