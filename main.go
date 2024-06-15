package main

import (
	"fmt"
	"mmih/rabbit"
)

func main() {
	fmt.Println("Starting mmih server")
	fmt.Println("Attempting to connect to RMQ server")
	var forever chan struct {};
	go rabbit.IntializeRMQClient();
	<-forever
}
