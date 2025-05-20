package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
)

func (node *Node) run() {
	electionTimeoutCh := make(chan electionTimer.ElectionTimeoutSignal)

	go func() {
		for {
			select {
			case <-electionTimeoutCh:
				// handle election timout: start a new election
				node.startElection()
			}
		}
	}()

	// start the election timer
	node.electionTimer.StartReq <- electionTimeoutCh
}
