package main

import (
	"LeaderElectionGo/leaderElection"
	"fmt"
)

func main() {

	// number of nodes
	fmt.Println("How many nodes do you want to run in the cluster?")

	var numNodes int
	fmt.Scanln(&numNodes)

	leaderElection.NewNode("1", "localhost:50051", nil)

	for {
	}
}
