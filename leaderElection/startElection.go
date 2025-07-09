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
	responseCh := make(chan bool)
	node.myVote.SetVoteReq <- myVote.SetVoteSignal{
		Vote:       node.ID,
		Term:       term,
		ResponseCh: responseCh,
	}
	if success := <-responseCh; !success {
		node.state.FollowerCh <- state.FollowerSignal{}
	}

	becomeLeaderCh := make(chan bool)
	node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
		Term:           term,
		VoterID:        node.ID,
		BecomeLeaderCh: becomeLeaderCh,
	}

	// checks if we can become the leader (cluster of size <= 2)
	if becomeLeaderFlag := <-becomeLeaderCh; becomeLeaderFlag {
		// we can become the leader
		node.handleLeadership(term)
	}
	// else we did not get enough votes to become leader

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
