package leaderElection

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

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
			node.voteCount.AddVoteReq <- voteCount.AddVoteSignal{
				Term:    term,
				VoterID: nodeID,
			}
		} // else we do nothing (vote not granted)
	}
}
