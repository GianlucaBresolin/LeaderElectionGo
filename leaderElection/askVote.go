package leaderElection

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	"LeaderElectionGo/leaderElection/state"
	"LeaderElectionGo/leaderElection/voteCount"
	pb "LeaderElectionGo/leaderElection/voteRequestService"
)

const RETRY_DELAY = 20

func (node *Node) askVote(nodeID string, term int, conn *grpc.ClientConn) {
	client := pb.NewVoteRequestServiceClient(conn)

	req := &pb.VoteRequest{
		Term:        int32(term),
		CandidateId: node.ID,
	}

	successFlag := false
	for !successFlag {
		resp, err := client.VoteRequestGRPC(context.Background(), req)
		if err != nil {
			log.Printf("Error sending vote request to node %s: %v. Retrying...", nodeID, err)
			time.Sleep(RETRY_DELAY * time.Millisecond) // avoid busy looping
			continue
		}
		// the node responded
		successFlag = true
		if resp.Granted {
			// we got the vote
			becomeLeaderCh := make(chan bool)
			node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
				Term:           term,
				VoterID:        nodeID,
				BecomeLeaderCh: becomeLeaderCh,
			}

			if becomeLeaderFlag := <-becomeLeaderCh; becomeLeaderFlag {
				// become the leader
				node.state.LeaderCh <- state.LeaderSignal{}
				log.Printf("Node %s has become the leader for term %d", node.ID, term)
				return
			}
			// else we did not get enough votes to become leader
		}
		// else we do nothing (vote not granted)
	}
}
