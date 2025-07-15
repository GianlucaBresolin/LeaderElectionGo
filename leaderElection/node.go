package leaderElection

import (
	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/heartbeatTimer"
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/stopLeadershipSignal"
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

	configurationMap map[string]connectionData                      // map[nodeID]connectionData
	stopLeadershipCh chan stopLeadershipSignal.StopLeadershipSignal // channel to stop leadership
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
		heartbeatTimer:   heartbeatTimer.NewHeartbeatTimer(10),
		voteCount:        voteCount.NewVoteCount(addressMap),
		myVote:           myVote.NewMyVote(),
		currentLeader:    currentLeader.NewCurrentLeader(),
		configurationMap: configurationMap,
		stopLeadershipCh: nil, // will be set when the node becomes a leader
		CloseCh:          make(chan CloseSignal),
	}

	node.prepareServer()
	node.prepareConnections()

	go node.run()

	return node
}

func (node *Node) run() {
	electionTimeoutCh := make(chan electionTimer.ElectionTimeoutSignal)
	becomeLeaderCh := make(chan becomeLeaderSignal)

	go func() {
		for {
			select {
			case <-electionTimeoutCh:
				// handle election timout: start a new election
				go node.handleElection(becomeLeaderCh)
			case signal := <-becomeLeaderCh:
				go node.handleLeadership(signal.term)
			case <-node.CloseCh:
				// handle close signal:
				// close connections to other nodes
				node.closeConnections()
			}
		}
	}()

	// start the election timer
	node.electionTimer.StartReq <- electionTimer.StartSignal{
		ResponseCh: electionTimeoutCh,
	}
}
