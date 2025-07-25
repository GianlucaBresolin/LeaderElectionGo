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

const RETRY_DELAY = 20 // milliseconds

type becomeLeaderSignal struct {
	term int
}

func (node *Node) handleElection(becomeLeaderCh chan becomeLeaderSignal) {
	log.Println("NODE", node.ID, "START ELECTION")
	// reset the election timer to resolve split-votes
	node.electionTimer.ResetReq <- electionTimer.ResetSignal{}

	// increment the current term
	termCh := make(chan int)
	node.currentTerm.IncReq <- term.IncrementSignal{
		ResponseCh: termCh,
	}
	// get the incremented term
	term := <-termCh

	// become a candidate
	node.state.CandidateCh <- state.CandidateSignal{
		Term: term,
	}

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
		node.state.FollowerCh <- state.FollowerSignal{
			HeartbeatTimerRef: node.heartbeatTimer,
			ElectionTimerRef:  node.electionTimer,
			StopLeadershipCh:  node.stopLeadershipCh,
			Term:              term,
		}
		return
	}

	becomeLeaderFlagCh := make(chan bool)
	node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
		Term:           term,
		VoterID:        node.ID,
		BecomeLeaderCh: becomeLeaderFlagCh,
	}

	// checks if we can become the leader (cluster of size <= 2)
	if becomeLeaderFlag := <-becomeLeaderFlagCh; becomeLeaderFlag {
		// we can become the leader
		becomeLeaderCh <- becomeLeaderSignal{
			term: term,
		}
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
		go node.askVote(nodeID, term, becomeLeaderCh, conn)
	}
}

func (node *Node) askVote(nodeID utils.NodeID, term int, becomeLeaderCh chan becomeLeaderSignal, conn *grpc.ClientConn) {
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
			time.Sleep(RETRY_DELAY * time.Millisecond) // avoid busy looping
			continue                                   // retry sending vote request
		}
		// the node responded
		successFlag = true
		if resp.Granted {
			// we got the vote
			becomeLeaderFlagCh := make(chan bool)
			node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
				Term:           term,
				VoterID:        nodeID,
				BecomeLeaderCh: becomeLeaderFlagCh,
			}

			if becomeLeaderFlag := <-becomeLeaderFlagCh; becomeLeaderFlag {
				// become the leader
				becomeLeaderCh <- becomeLeaderSignal{
					term: term,
				}
				return
			}
			// else we did not get enough votes to become leader
		}
		// else we do nothing (vote not granted)
	}
}
