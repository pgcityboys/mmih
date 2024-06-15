package matching

import (
	"log"
	"mmih/rabbit"
)

func InitializeMatchingClient() {
	go func() {
		for matchRequest := range rabbit.MatchRequests {
			log.Printf("Received match request: %s", matchRequest.String())
		}
	}()
}
