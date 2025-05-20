package leaderElection

import (
	"LeaderElectionGo/leaderElection/myVote"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/voteCount"
)

func (node *Node) startElection() {
	// become a candidate
	node.state.CandidateCh <- state.CandidateSignal{}

	// increment the current term
	node.currentTerm.IncReq <- term.IncrementSignal{}

	// reset the vote count
	node.voteCount.ResetReq <- voteCount.ResetSignal{}

	// vote for myself
	node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{}
	node.myVote.SetVoteReq <- myVote.SetVoteArg{
		Vote: node.ID,
	}

	// send vote request to all other nodes
	// TODO

}
