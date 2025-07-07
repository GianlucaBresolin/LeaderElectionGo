package leaderElection

import (
	"LeaderElectionGo/leaderElection/heartbeatTimer"
	"LeaderElectionGo/leaderElection/state"
	"log"
)

func (node *Node) handleLeadership(term int) {

	responseCh := make(chan bool)
	node.state.LeaderCh <- state.LeaderSignal{
		Term:       term,
		ResponseCh: responseCh,
	}

	if success := <-responseCh; success {
		// successfully set the state to leader
		log.Printf("Node %s has become the leader for term %d", node.ID, term)

		// start the heartbeat timer
		heartbeatTimeoutCh := make(chan heartbeatTimer.HeartbeatTimeoutSignal)
		node.heartbeatTimer.StartReq <- heartbeatTimer.StartSignal{
			ResponseCh: heartbeatTimeoutCh,
		}

		go func() {
			for {
				select {
				case <-heartbeatTimeoutCh:
					// handle heartbeat timeout: send heartbeats to all followers
					node.sendHeartbeats(term)
					node.heartbeatTimer.ResetReq <- heartbeatTimer.ResetSignal{}
				}
			}
		}()
	}
	// else, another handleLeadership is currently in progress or the term has changed
}
