package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
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

		// stop the election timer
		node.electionTimer.StopReq <- electionTimer.StopSignal{}

		// provide heartbeats
		node.sendHeartbeats(term)

		// start the heartbeat timer
		heartbeatTimeoutCh := make(chan heartbeatTimer.HeartbeatTimeoutSignal)
		node.heartbeatTimer.StartReq <- heartbeatTimer.StartSignal{
			ResponseCh: heartbeatTimeoutCh,
		}

		// i := 0

		go func() {
			for {
				select {
				case <-heartbeatTimeoutCh:
					// i++
					// if i == 100 {
					// 	// sleep for a while to allow other leaders
					// 	time.Sleep(200 * time.Millisecond)
					// 	i = 0 // reset the counter
					// }
					// handle heartbeat timeout: send heartbeats to all followers
					node.heartbeatTimer.ResetReq <- heartbeatTimer.ResetSignal{}
					node.sendHeartbeats(term)
				}
			}
		}()
	}
	// else, another handleLeadership is currently in progress or the term has changed
}
