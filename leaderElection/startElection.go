package leaderElection

import (
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/voteCount"
	"log"
)

func (node *Node) startElection() {
	log.Println("NODE", node.ID, "START ELECTION")
	// become a candidate
	node.state.CandidateCh <- state.CandidateSignal{}

	// increment the current term
	termCh := make(chan int)
	node.currentTerm.IncReq <- term.IncrementSignal{
		ResponseCh: termCh,
	}
	// get the incremented term
	term := <-termCh

	// reset the vote count
	node.voteCount.ResetReq <- voteCount.ResetSignal{
		Term: term,
	}

	// vote for myself
	node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
		Term:    term,
		VoterID: node.ID,
	}
	node.myVote.SetVoteReq <- myVote.SetVoteSignal{
		Vote: node.ID,
	}

	// send vote request to all other nodes
	for nodeID, connData := range node.configurationMap {
		if nodeID == node.ID {
			continue
		}

		conn := connData.connection
		if conn == nil {
			log.Printf("Error node %s: connection to node %s is nil.", node.ID, nodeID)
			continue
		}

		// send vote request (in parallel)
		go node.askVote(nodeID, term, conn)
	}
}
