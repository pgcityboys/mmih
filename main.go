package main

import (
	"fmt"
	"mmih/matching"
	"mmih/rabbit"
)

func main() {
	fmt.Println("Starting mmih server")
	fmt.Println("Attempting to connect to RMQ server")
	var forever chan struct {};
	go rabbit.IntializeRMQClient();
	go matching.InitializeMatchingClient();
	<-forever
}
