package matcher

import (
	"log"
	"mmih/rabbit"
	"mmih/storage"
)

func InitializeMatching() {
	var forever chan struct{}
	go handleNewRoom()
	go handleJoinRoom()
	<-forever
}

func handleNewRoom() {
	for request := range rabbit.NewRoomChannel {
		log.Printf("Received: %s", request.String())
		users := []string{request.UserId}
		data := storage.RoomInfo{Users: users, MaxUsers: int(request.MaxUsers), Description: request.Description, Category: request.Category}
		storage.CreateEmptyRoom(&data)
	}
	//TODO - send join info
}

func handleJoinRoom() {
	for request := range rabbit.JoinRoomChannel {
		log.Printf("Received match request: %s", request.String())
		err := storage.JoinRoom(request.UserId, request.RoomId);
		if err != nil {
			log.Println("system has fallen from rowerek")
		}
		info, err := storage.GetRoom(request.RoomId);
		log.Printf("There are %d users after joining", len(info.Users))
	}
}
