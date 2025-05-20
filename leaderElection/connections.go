package leaderElection

import (
	"log"
	"net"

	pb "LeaderElectionGo/leaderElection/voteRequestService"

	"google.golang.org/grpc"
)

func (node *Node) prepareConnections() {

	listener, err := net.Listen("tcp", node.address)
	if err != nil {
		log.Fatalf("Error starting listener: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterVoteRequestServiceServer(grpcServer, node)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting gRPC server: %v", err)
	}
	log.Printf("Node %s is listening on %s", node.ID, node.address)
}
