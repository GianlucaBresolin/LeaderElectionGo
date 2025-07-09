package leaderElection

import (
	"context"
	"log"

	pb "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	"LeaderElectionGo/leaderElection/state"

	"google.golang.org/grpc"
)

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
		Leader: node.ID,
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
				HeartbeatTimerRef: node.heartbeatTimer,
				ElectionTimerRef:  node.electionTimer,
			}
		}
		// else nothing to do, heartbeat was received successfully
	}
}
