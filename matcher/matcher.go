package matcher

import (
	"log"
	"mmih/messages"
	"mmih/rabbit"
	"mmih/storage"
)

func InitializeMatching() {
	var forever chan struct{}
	go handleNewRoom()
	go handleJoinRoom()
	go handleRoomInfo()
	go handleRoomLeave()
	<-forever
}

func handleNewRoom() {
	for request := range rabbit.NewRoomChannel {
		users := []string{request.UserId}
		data := storage.RoomInfo{Users: users, MaxUsers: int(request.MaxUsers), Description: request.Description, Category: request.Category}
		storage.CreateEmptyRoom(&data)
	}
	//TODO - send join info
}

func handleJoinRoom() {
	for request := range rabbit.JoinRoomChannel {
		err := storage.JoinRoom(request.UserId, request.RoomId);
		if err != nil {
			log.Println("system has fallen from rowerek")
		}
		info, err := storage.GetRoom(request.RoomId);
		log.Printf("There are %d users after joining", len(info.Users))
	}
	// TODO - send join info
}

func handleRoomInfo() {
	for request := range rabbit.RoomInfoChannel {
		ids, _ := storage.CategoryRooms(request);
		var info []*messages.RoomInfo;
		for _, room := range ids {
			r, _ := storage.GetRoom(room)
			info = append(info, &messages.RoomInfo{RoomId: room, MaxUsers: int32(r.MaxUsers), Description: r.Description, CurrentUsers: int32(len(r.Users))})
		}
		res := messages.CategoryInfo{Category: request, Rooms: info}
		log.Println(res)
	}
	// TODO - send room info
}

func handleRoomLeave() {
	for request := range rabbit.RoomLeaveChannel {
		log.Println("Received request for room leave")
		storage.LeaveRoom(request.UserId, request.RoomId)
		// todo - send room info
	}
}
