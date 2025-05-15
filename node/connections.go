package node

import (
	"log"
	"net"

	pb "LeaderElectionGo/node/voteRequestService"

	"google.golang.org/grpc"
)

func (self *Node) PrepareConnections() {

	listener, err := net.Listen("tcp", self.address)
	if err != nil {
		log.Fatalf("Error starting listener: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterVoteRequestServiceServer(grpcServer, self)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error starting gRPC server: %v", err)
	}
	log.Printf("Node %s is listening on %s", self.ID, self.address)
}
