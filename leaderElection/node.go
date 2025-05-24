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

func NewNode(id string, address string, addressMap map[string]string) *Node {
	// build the configurationMap from the addressMap
	configurationMap := make(map[string]connectionData)
	for nodeID, address := range addressMap {
		configurationMap[nodeID] = connectionData{
			address:    address,
			connection: nil,
		}
	}

	node := &Node{
		ID:               id,
		address:          address,
		state:            state.NewState(),
		currentTerm:      term.NewTerm(),
		electionTimer:    electionTimer.NewElectionTimer(150, 300),
		voteCount:        voteCount.NewVoteCount(addressMap),
		myVote:           myVote.NewMyVote(),
		configurationMap: configurationMap,
		CloseCh:          make(chan CloseSignal),
	}

	node.prepareServer()
	node.prepareConnections()

	go node.run()

	return node
}
