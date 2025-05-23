package main

import (
	"LeaderElectionGo/leaderElection"
)

func main() {
	addressMap := make(map[string]string)
	addressMap["1"] = "localhost:50051"
	addressMap["2"] = "localhost:50052"
	addressMap["3"] = "localhost:50053"

	leaderElection.NewNode("1", "localhost:50051", addressMap)
	leaderElection.NewNode("2", "localhost:50052", addressMap)
	leaderElection.NewNode("3", "localhost:50053", addressMap)

	for {
	}
}
