package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/internalUtils"
	pb "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	"LeaderElectionGo/leaderElection/state"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const HEARTBEAT_TIMEOUT = 50 * time.Millisecond

func (node *Node) handleLeadership(term int) {

	responseCh := make(chan bool)
	node.state.LeaderCh <- state.LeaderSignal{
		Term:       term,
		ResponseCh: responseCh,
	}

	if success := <-responseCh; success {
		// stop the election timer
		node.electionTimer.StopReq <- electionTimer.StopSignal{}

		// successfully set the state to leader
		log.Printf("Node %s has become the leader for term %d", node.ID, term)

		// set the stopLeadershipCh to allow stopping the leadership
		if node.stopLeadershipCh != nil {
			log.Fatal("Error: stopLeadershipCh is not nil, it should be nil before starting leadership, incongruce state.")
		} else {
			node.stopLeadershipCh = make(chan internalUtils.StopLeadershipSignal)
		}

		// provide heartbeats
		node.sendHeartbeats(term)

		// i := 0
		for {
			select {
			case <-time.After(HEARTBEAT_TIMEOUT):
				// i++
				// if i == 10 {
				// 	// sleep for a while to allow other leaders
				// 	time.Sleep(200 * time.Millisecond)
				// 	i = 0 // reset the counter
				// }
				// handle heartbeat timeout: send heartbeats to all followers
				node.sendHeartbeats(term)
			case <-node.stopLeadershipCh:
				log.Println("Node", node.ID, "stopping leadership for term", term)
				node.stopLeadershipCh = nil // reset the channel to nil
				return
			}
		}
	}
	// else, another handleLeadership is currently in progress or the term has changed, exit
}

func (node *Node) sendHeartbeats(term int) {
	// send heartbeats to all followers
	for nodeID, connData := range node.configurationMap {
		if nodeID == node.ID {
			continue // skip self
		}

		conn := connData.connection
		if conn == nil {
			log.Printf("Error node %s: connection to node %s is nil.", node.ID, nodeID)
			continue
		}

		// send heartbeat (in parallel)
		go node.sendHeartbeat(term, conn)
	}
}

func (node *Node) sendHeartbeat(term int, conn *grpc.ClientConn) {
	client := pb.NewHeartbeatServiceClient(conn)

	req := &pb.HeartbeatRequest{
		Term:   int32(term),
		Leader: string(node.ID),
	}

	successFlag := false
	for !successFlag {
		resp, err := client.HeartbeatRequestGRPC(context.Background(), req)
		if err != nil {
			log.Printf("Error sending heartbeat to node: %v. Retrying...", err)
			continue // retry sending heartbeat
		}
		// the node responded
		successFlag = true
		if !resp.Success {
			// heartbeat was not successful, revert to follower state
			node.state.FollowerCh <- state.FollowerSignal{
				ElectionTimerRef: node.electionTimer,
				StopLeadershipCh: node.stopLeadershipCh,
				Term:             term,
			}
		}
		// else nothing to do, heartbeat was received successfully
	}
}
