package main

import (
	"fmt"
)

func main() {

	// number of nodes
	fmt.Println("How many nodes do you want to run in the cluster?")

	var numNodes int
	fmt.Scanln(&numNodes)
}
