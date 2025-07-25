package main

import (
	"LeaderElectionGo/leaderElection"
	"log"
	"os"
)

func main() {
	log.Println("Starting Node...")

	// build the configuration map from command line arguments
	peers := make(map[leaderElection.NodeID]leaderElection.Address)
	if len(os.Args) > 2 {
		for i := 3; i < len(os.Args); i += 2 {
			peers[leaderElection.NodeID(os.Args[i])] = leaderElection.Address(os.Args[i+1])
		}
	}

	// build the node
	leaderElection.NewNode(leaderElection.NodeID(os.Args[1]), leaderElection.Address(os.Args[2]), peers)

	for {
	}
}
