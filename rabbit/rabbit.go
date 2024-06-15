package rabbit;

import (
	"context"
	"fmt"
	"log"
	"mmih/messages"
	"mmih/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

var connection *amqp.Connection;
var channel *amqp.Channel;

const TIMEOUT time.Duration = 5;
const RETRIES int = 5;
const SEND_TIMEOUT time.Duration = 10;

var NewRoomChannel chan messages.NewRoomRequest = make(chan messages.NewRoomRequest);
var JoinRoomChannel chan messages.MatchRequest = make(chan messages.MatchRequest);
var RoomInfoChannel chan messages.CategoryInfoRequest = make(chan messages.CategoryInfoRequest);
var RoomLeaveChannel chan messages.MatchRequest = make(chan messages.MatchRequest);
var ChatInChannel chan messages.ChatIn = make(chan messages.ChatIn);

func IntializeRMQClient() {
	rabbitAddress := utils.EnvWithDefaults("RABBITMQ_ADDRESS", "localhost")
	rabbitUser := utils.EnvWithDefaults("RABBITMQ_USER", "guest")
	rabbitPassword := utils.EnvWithDefaults("RABBITMQ_PASSWORD", "guest")
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:5672/", rabbitUser, rabbitPassword,rabbitAddress)
	for range RETRIES {
		conn, err := amqp.Dial(connectionString)
		if err != nil {
			log.Println("Could not connect to rmq, retrying")
			time.Sleep(TIMEOUT * time.Second)
		} else { // Initialize client
			connection = conn;
			defer connection.Close();
			err := establishChannel();
			if err != nil {
				continue;
			}
			log.Println("Connected to rmq instance")
			var forever chan struct {};
			go handleMatchRequests()
			go handleNewRoomRequests()
			go handleRoomInfo()
			go handleRoomLeave()
			<-forever;
		}
	}
	log.Println("RMQ client: waiting for incoming messages")
}

func establishChannel() error {
	c, err := connection.Channel();
	if err != nil {
		log.Println("ERROR: Could not create channel")
		return amqp.Error{};
	}
	channel = c;
	// Receive messages from web app
	channel.ExchangeDeclare("web", "direct", false, false, false, false, nil);
	channel.QueueDeclare("match_req", false, false, false, false, nil);
	channel.QueueDeclare("chat_req", false, false, false, false, nil)
	channel.QueueDeclare("rooms_req", false, false, false, false, nil)
	channel.QueueDeclare("rooms_new", false, false, false, false, nil)
	channel.QueueDeclare("leave_room", false, false, false, false, nil)
	channel.QueueBind("match_req", "match_req", "web", false, nil);
	channel.QueueBind("chat_req", "chat_req", "web", false, nil);
	channel.QueueBind("rooms_req", "rooms_req", "web", false, nil);
	channel.QueueBind("rooms_new", "rooms_new", "web", false, nil);
	channel.QueueBind("leave_room", "leave_room", "web", false, nil);
	// Send out messages to topic exchange
	channel.ExchangeDeclare("matchmaking", "topic", false, false, false, false, nil);
	channel.QueueDeclare("match_res", false, false, false, false, nil);
	channel.QueueDeclare("chat_notify", false, false, false, false, nil)
	channel.QueueDeclare("room_info", false, false, false, false, nil)
	channel.QueueDeclare("room_leave", false, false, false, false, nil)
	channel.QueueBind("match_res", "match_res", "matchmaking", false, nil);
	channel.QueueBind("chat_notify", "chat_notify", "matchmaking", false, nil);
	channel.QueueBind("room_info", "room_info", "matchmaking", false, nil);
	channel.QueueBind("room_leave", "room_leave", "matchmaking", false, nil);
	return nil;
}

func ensureChannelHealth() error {
	if channel == nil || channel.IsClosed() {
		return establishChannel();
	}
	return nil;
}

// Handlers

func handleMatchRequests() {
	msgs, _ := channel.ConsumeWithContext(context.Background(), "match_req", "mmih", true, false, false, false, nil)
	for msg := range msgs {
		var request messages.MatchRequest;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}
		JoinRoomChannel<-request
	}
}

func handleNewRoomRequests() {
	newRoomChannel, _ := connection.Channel();
	msgs, _ := newRoomChannel.ConsumeWithContext(context.Background(), "rooms_new", "mmih", true, false, false, false, nil);
	for msg := range msgs {
		var request messages.NewRoomRequest;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}
		NewRoomChannel<-request
	}
}

func handleRoomInfo() {
	roomInfoChannel, _ := connection.Channel();
	msgs, _ := roomInfoChannel.ConsumeWithContext(context.Background(), "rooms_req", "mmih", true, false, false, false, nil);
	for msg := range msgs {
		var request messages.CategoryInfoRequest;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}

		RoomInfoChannel<-request
	}
}

func handleRoomLeave() {
	roomLeaveChannel, _ := connection.Channel();
	msgs, _ := roomLeaveChannel.ConsumeWithContext(context.Background(), "leave_room", "mmih", true, false, false, false, nil);
	for msg := range msgs {
		var request messages.MatchRequest;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}
		RoomLeaveChannel<-request
	}
}

func handleChatIn() {
	chatChannel, _ := connection.Channel();
	msgs, _ := chatChannel.ConsumeWithContext(context.Background(), "chat_req", "mmih", true, false, false, false, nil);
	for msg := range msgs {
		var request messages.ChatIn;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}
		ChatInChannel<-request
	}

}

// Send messages
func SendRoomUpdate(data messages.MatchConfirmation) {
	binaryData, _ := proto.Marshal(&data);
	channel.Publish("matchmaking", "match_res", false, false, amqp.Publishing{Body: binaryData})
}

func SendCategoryInfo(data messages.CategoryInfo) {
	binaryData, _ := proto.Marshal(&data);
	channel.Publish("matchmaking", "room_info", false, false, amqp.Publishing{Body: binaryData})
}

func SendRoomLeave(data messages.LeaveRoomInfo) {
	binaryData, _ := proto.Marshal(&data);
	channel.Publish("matchmaking", "room_leave", false, false, amqp.Publishing{Body: binaryData})
}

func SendChatOut(data messages.ChatOut) {
	binaryData, _ := proto.Marshal(&data);
	channel.Publish("matchmaking", "chat_notify", false, false, amqp.Publishing{Body: binaryData})
}
