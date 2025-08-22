package leaderElection

import (
	"LeaderElectionGo/leaderElection/electionTimer"
	"LeaderElectionGo/leaderElection/myVote"
	pb "LeaderElectionGo/leaderElection/services/voteRequest/voteRequestService"
	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/term"
	"LeaderElectionGo/leaderElection/utils"
	"LeaderElectionGo/leaderElection/voteCount"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const RETRY_DELAY = 20 * time.Millisecond

func (node *Node) handleElection(electionTerm int, becomeLeaderCh chan voteCount.BecomeLeaderSignal) {
	getStateResponseCh := make(chan state.GetStateResponse)
	node.state.GetState(state.GetStateSignal{
		ResponseCh: getStateResponseCh,
	})
	if response := <-getStateResponseCh; response.State == "leader" {
		// node is already a leader, no need to start another election
		return
	}

	log.Println("NODE", node.ID, "START ELECTION")

	// set the term for the election
	successSetTermCh := make(chan bool)
	node.currentTerm.SetTermReq <- term.SetTermSignal{
		Value:      electionTerm,
		ResponseCh: successSetTermCh,
	}
	if success := <-successSetTermCh; !success {
		// failed to set term, do not proceed
		return
	}
	// else, term is set successfully: valid election

	// reset the election timer to resolve possible split-votes
	node.electionTimer.ResetReq <- electionTimer.ResetSignal{
		Term: electionTerm,
	}

	// become a candidate
	successCandidateCh := make(chan bool)
	node.state.CandidateCh <- state.CandidateSignal{
		Term:       electionTerm,
		ResponseCh: successCandidateCh,
	}
	if success := <-successCandidateCh; !success {
		// failed to become a candidate, quit election
		return
	}

	// vote for myself
	responseCh := make(chan bool)
	node.myVote.SetVoteReq <- myVote.SetVoteSignal{
		Vote:       node.ID,
		Term:       electionTerm,
		ResponseCh: responseCh,
	}
	if success := <-responseCh; success {
		// count our vote
		node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
			Term:           electionTerm,
			VoterID:        node.ID,
			BecomeLeaderCh: becomeLeaderCh,
		}
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
		go node.askVote(nodeID, electionTerm, becomeLeaderCh, conn)
	}
}

func (node *Node) askVote(nodeID utils.NodeID, term int, becomeLeaderCh chan voteCount.BecomeLeaderSignal, conn *grpc.ClientConn) {
	client := pb.NewVoteRequestServiceClient(conn)

	req := &pb.VoteRequest{
		Term:        int32(term),
		CandidateId: string(node.ID),
	}

	successFlag := false
	for !successFlag {
		resp, err := client.VoteRequestGRPC(context.Background(), req)
		if err != nil {
			log.Printf("Error sending vote request to node %s: %v. Retrying...", nodeID, err)
			// avoid busy looping
			time.Sleep(RETRY_DELAY)
			// retry sending vote request
			continue
		}
		// the node responded
		successFlag = true
		if resp.Granted {
			// we got the vote: count it
			node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
				Term:           term,
				VoterID:        nodeID,
				BecomeLeaderCh: becomeLeaderCh,
			}
		}
		// else we do nothing (vote not granted)
	}
}
