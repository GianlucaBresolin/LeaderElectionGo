package leaderElection

import (
	"context"
	"log"

	"LeaderElectionGo/leaderElection/allowVotes"
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/myVote"
	pb "LeaderElectionGo/leaderElection/services/voteRequest/voteRequestService"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/utils"
)

func (node *Node) VoteRequestGRPC(ctx context.Context, req *pb.VoteRequest) (*pb.VoteResponse, error) {
	// response
	voteResponse := &pb.VoteResponse{
		// the vote is not granted by default
		Granted: false,
	}

	getAllowVotesResponseCh := make(chan bool)
	node.allowVotes.GetAllowVotesCh <- allowVotes.GetAllowVotesCh{
		ResponseCh: getAllowVotesResponseCh,
	}
	if allowed := <-getAllowVotesResponseCh; allowed {
		// read the current term
		currentTermCh := make(chan int)
		node.currentTerm.GetTermReq <- term.GetTermSignal{
			ResponseCh: currentTermCh,
		}
		currentTerm := <-currentTermCh

		switch {
		case int(req.Term) < currentTerm:
			// the term is not valid, not grant the vote (default)
			return voteResponse, nil
		case int(req.Term) > currentTerm:
			// try to update the term
			successSetTermCh := make(chan bool)
			node.currentTerm.SetTermReq <- term.SetTermSignal{
				Value:      int(req.Term),
				ResponseCh: successSetTermCh,
			}
			if success := <-successSetTermCh; success {
				// revert to follower state
				successRevertToFollowerCh := make(chan bool)
				node.state.FollowerCh <- state.FollowerSignal{
					Term:       int(req.Term),
					ResponseCh: successRevertToFollowerCh,
				}
				if success := <-successRevertToFollowerCh; success {
					// reset the election timer
					node.electionTimer.ResetReq <- electionTimer.ResetSignal{
						Term: int(req.Term),
					}
				}
			}
		default:
			// the term matches, proceed to set our vote
		}

		responseCh := make(chan bool)
		node.myVote.SetVoteReq <- myVote.SetVoteSignal{
			Vote:       utils.NodeID(req.CandidateId),
			Term:       int(req.Term),
			ResponseCh: responseCh,
		}
		if success := <-responseCh; success {
			voteResponse.Granted = true
		}
		// else do nothing (vote not granted)
	}
	log.Println("NODE", node.ID, "VOTE REQUEST FROM NODE", req.CandidateId, "FOR TERM", req.Term, "GRANTED:", voteResponse.Granted)
	return voteResponse, nil
}
