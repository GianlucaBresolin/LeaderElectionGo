package leaderElection

import (
	"LeaderElectionGo/leaderElection/allowVotes"
	"LeaderElectionGo/leaderElection/electionTimer"
	pb "LeaderElectionGo/leaderElection/services/heartbeatRequest/heartbeatRequestService"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const HEARTBEAT_TIMEOUT = 50 * time.Millisecond

func (node *Node) handleLeadership(leadershipTerm int) {
	responseCh := make(chan bool)
	node.state.LeaderCh <- state.LeaderSignal{
		Term:       leadershipTerm,
		ResponseCh: responseCh,
	}
	if success := <-responseCh; success {
		// stop the election timer
		stopElectionTimerCh := make(chan bool)
		node.electionTimer.StopReq <- electionTimer.StopSignal{
			Term:       leadershipTerm,
			ResponseCh: stopElectionTimerCh,
		}
		if success := <-stopElectionTimerCh; !success {
			// failed to stop the election timer, do not proceed
			return
		}

		// stop allowing votes
		node.allowVotes.StopCh <- allowVotes.StopSignal{}

		// successfully set the state to leader
		log.Printf("Node %s has become the leader for term %d", node.ID, leadershipTerm)

		// provide heartbeats
		node.sendHeartbeats(leadershipTerm)

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

				// check if we are still the leader
				getStateResponseCh := make(chan state.GetStateResponse)
				node.state.GetStateCh <- state.GetStateSignal{
					ResponseCh: getStateResponseCh,
				}
				if response := <-getStateResponseCh; !(response.State == "leader" && response.Term == leadershipTerm) {
					// quit leadership: we are no longer the leader
					log.Println("Node", node.ID, "stopping leadership for term", leadershipTerm)
					// restart the allow votes mechanism
					node.allowVotes.RestartCh <- allowVotes.RestartSignal{}
					return
				}
				// else send heartbeats to all followers
				node.sendHeartbeats(leadershipTerm)
			}
		}
	}
	// else, another handleLeadership is currently in progress or the term has changed, exit
}

func (node *Node) sendHeartbeats(leadershipTerm int) {
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
		go node.sendHeartbeat(leadershipTerm, conn)
	}
}

func (node *Node) sendHeartbeat(leadershipTerm int, conn *grpc.ClientConn) {
	client := pb.NewHeartbeatServiceClient(conn)

	req := &pb.HeartbeatRequest{
		Term:   int32(leadershipTerm),
		Leader: string(node.ID),
	}

	successFlag := false
	for !successFlag {
		resp, err := client.HeartbeatRequestGRPC(context.Background(), req)
		if err != nil {
			log.Printf("Error sending heartbeat to node: %v. Retrying...", err)
			// retry sending heartbeat
			continue
		}
		successFlag = true
		if !resp.Success {
			// heartbeat was not successful, revert to follower state
			successRevertToFollowerCh := make(chan bool)
			node.state.FollowerCh <- state.FollowerSignal{
				Term:       int(resp.Term),
				ResponseCh: successRevertToFollowerCh,
			}
			if success := <-successRevertToFollowerCh; success {
				// reset the election timer
				node.electionTimer.ResetReq <- electionTimer.ResetSignal{
					Term: int(resp.Term),
				}
			}
			// try to update our term
			node.currentTerm.SetTermReq <- term.SetTermSignal{
				Value: int(resp.Term),
			}
		}
		// else nothing to do, heartbeat was received successfully
	}
}
