package leaderElection

import (
	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/heartbeatTimer"
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/voteCount"

	pb2 "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	pb1 "LeaderElectionGo/leaderElection/services/voteRequest/voteRequestService"
)

type Node struct {
	ID             string
	address        string
	state          *state.State
	currentTerm    *term.Term
	electionTimer  *electionTimer.ElectionTimer
	heartbeatTimer *heartbeatTimer.HeartbeatTimer
	voteCount      *voteCount.VoteCount
	myVote         *myVote.MyVote
	currentLeader  *currentLeader.CurrentLeader

	configurationMap map[string]connectionData // map[nodeID]connectionData
	CloseCh          chan CloseSignal

	pb1.UnimplementedVoteRequestServiceServer
	pb2.UnimplementedHeartbeatServiceServer
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
		heartbeatTimer:   heartbeatTimer.NewHeartbeatTimer(50),
		voteCount:        voteCount.NewVoteCount(addressMap),
		myVote:           myVote.NewMyVote(),
		currentLeader:    currentLeader.NewCurrentLeader(),
		configurationMap: configurationMap,
		CloseCh:          make(chan CloseSignal),
	}

	node.prepareServer()
	node.prepareConnections()

	go node.run()

	return node
}
