package main

import (
	"fmt"
	"mmih/matcher"
	"mmih/rabbit"
	"mmih/storage"
	"sync"
)

var wg sync.WaitGroup;

func main() {
	fmt.Println("Starting mmih server")
	fmt.Println("Attempting to connect to RMQ server")
	wg.Add(1)
	go rabbit.IntializeRMQClient();
	go storage.InitializeRedisConnection();
	go matcher.InitializeMatching();
	wg.Wait()
}
