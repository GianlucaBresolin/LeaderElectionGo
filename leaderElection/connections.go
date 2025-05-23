package leaderElection

import (
	"log"
	"net"
	"time"

	pb "LeaderElectionGo/leaderElection/voteRequestService"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type connectionData struct {
	address    string
	connection *grpc.ClientConn
}

type CloseSignal struct{}

func (node *Node) prepareServer() {
	listener, err := net.Listen("tcp", node.address)
	if err != nil {
		log.Fatalf("Error starting listener: %v.", err)
	}

	grpcServer := grpc.NewServer()

	// voteRequestServiceServer
	pb.RegisterVoteRequestServiceServer(grpcServer, node)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Error %s serving gRPC server: %v.", node.ID, err)
		}
	}()
	log.Printf("Node %s is listening on %s.", node.ID, node.address)
}

func (node *Node) prepareConnections() {
	for nodeID, connData := range node.configurationMap {
		if nodeID == node.ID {
			continue
		}

		successFlag := false
		for !successFlag {
			conn, err := grpc.NewClient(connData.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to connect to node %s: %v. Retrying...", nodeID, err)
				// avoid busy looping
				time.Sleep(time.Second)
			}
			// Successfully connected
			successFlag = true
			log.Printf("Node %s connected to node %s.", node.ID, nodeID)

			node.configurationMap[nodeID] = connectionData{
				address:    connData.address,
				connection: conn,
			}
		}
	}
}

func (node *Node) closeConnections() {
	for nodeID, connData := range node.configurationMap {
		if nodeID == node.ID {
			continue
		}

		closeFlag := false
		for !closeFlag {
			if err := connData.connection.Close(); err != nil {
				log.Printf("Failed to close connection to node %s: %v.", nodeID, err)
				continue
			}
			// Successfully closed the connection
			closeFlag = true
			log.Printf("Node %s closed connection to node %s.", node.ID, nodeID)
		}
	}
}
