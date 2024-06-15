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
		id := storage.CreateEmptyRoom(&data)
		payload := messages.MatchConfirmation{Users: users, RoomId: id, Category: request.Category}
		rabbit.SendRoomUpdate(payload)
	}
}

func handleJoinRoom() {
	for request := range rabbit.JoinRoomChannel {
		err := storage.JoinRoom(request.UserId, request.RoomId);
		if err != nil {
			log.Println("system has fallen from rowerek")
		}
		info, err := storage.GetRoom(request.RoomId);
		payload := messages.MatchConfirmation{Users: info.Users, RoomId: request.RoomId, Category: info.Category}
		rabbit.SendRoomUpdate(payload)
	}
}

func handleRoomInfo() {
	for request := range rabbit.RoomInfoChannel {
		ids, _ := storage.CategoryRooms(request.Category);
		var info []*messages.RoomInfo;
		for _, room := range ids {
			r, _ := storage.GetRoom(room)
			info = append(info, &messages.RoomInfo{RoomId: room, MaxUsers: int32(r.MaxUsers), Description: r.Description, CurrentUsers: int32(len(r.Users))})
		}
		res := messages.CategoryInfo{Category: request.Category, Rooms: info, UserId: request.UserId}
		rabbit.SendCategoryInfo(res)
	}
}

func handleRoomLeave() {
	for request := range rabbit.RoomLeaveChannel {
		log.Println("Received request for room leave")
		user, others, err := storage.LeaveRoom(request.UserId, request.RoomId)
		if err != nil {
			continue
		}
		payload := messages.LeaveRoomInfo{UserId: user, Users: others}
		rabbit.SendRoomLeave(payload)
	}
}

func handleChatIn() {
	for request := range rabbit.ChatInChannel {
		log.Println("Received request for room leave")
		info, err := storage.GetRoom(request.RoomId)
		if err != nil {
			log.Println("Sent chat to a nonexistent room")
			continue
		}
		payload := messages.ChatOut{UserId: request.UserId, Content: request.Content, Users: info.Users, RoomId: request.RoomId}
		rabbit.SendChatOut(payload)
	}
}
