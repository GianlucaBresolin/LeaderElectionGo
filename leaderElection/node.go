package leaderElection

import (
	"LeaderElectionGo/leaderElection/currentLeader"
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/internalUtils"
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/utils"
	"LeaderElectionGo/leaderElection/voteCount"

	pb2 "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	pb1 "LeaderElectionGo/leaderElection/services/voteRequest/voteRequestService"
)

type NodeID string
type Address string

type Node struct {
	ID            utils.NodeID
	address       utils.Address
	state         *state.State
	currentTerm   *term.Term
	electionTimer *electionTimer.ElectionTimer
	voteCount     *voteCount.VoteCount
	myVote        *myVote.MyVote
	currentLeader *currentLeader.CurrentLeader

	configurationMap map[utils.NodeID]connectionData         // map[nodeID]connectionData
	stopLeadershipCh chan internalUtils.StopLeadershipSignal // channel to stop leadership
	CloseCh          chan CloseSignal

	pb1.UnimplementedVoteRequestServiceServer
	pb2.UnimplementedHeartbeatServiceServer
}

func NewNode(id NodeID, address Address, addressMap map[NodeID]Address) *Node {
	// build the configurationMap from the addressMap
	configurationMap := make(map[utils.NodeID]connectionData)
	for nodeID, address := range addressMap {
		configurationMap[utils.NodeID(nodeID)] = connectionData{
			address:    utils.Address(address),
			connection: nil,
		}
	}

	node := &Node{
		ID:            utils.NodeID(id),
		address:       utils.Address(address),
		state:         state.NewState(),
		currentTerm:   term.NewTerm(),
		electionTimer: electionTimer.NewElectionTimer(150, 300),
		voteCount: voteCount.NewVoteCount(
			func(addressMap map[NodeID]Address) []utils.NodeID {
				nodeList := make([]utils.NodeID, 0, len(configurationMap))
				for nodeID := range configurationMap {
					nodeList = append(nodeList, nodeID)
				}
				return nodeList
			}(addressMap)),
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
	becomeLeaderCh := make(chan internalUtils.BecomeLeaderSignal)

	go func() {
		for {
			select {
			case <-electionTimeoutCh:
				// handle election timout: start a new election
				go node.handleElection(becomeLeaderCh)
			case signal := <-becomeLeaderCh:
				go node.handleLeadership(signal.Term)
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
