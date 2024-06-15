package rabbit

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

var MatchRequests chan messages.MatchRequest = make(chan messages.MatchRequest);


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
			var forever chan struct {};
			err := establishChannel();
			if err != nil {
				continue;
			}
			log.Println("Connected to rmq instance")
			go handleMatchRequests()
			<-forever;
		}
	}
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
	channel.QueueBind("match_req", "match_req", "web", false, nil);
	channel.QueueBind("chat_req", "chat_req", "web", false, nil);
	// Send out messages to topic exchange
	channel.ExchangeDeclare("matchmaking", "topic", false, false, false, false, nil);
	channel.QueueDeclare("match_res", false, false, false, false, nil);
	channel.QueueDeclare("chat_notify", false, false, false, false, nil)
	channel.QueueBind("match_res", "match_res", "matchmaking", false, nil);
	channel.QueueBind("chat_notify", "chat_notify", "matchmaking", false, nil);
	return nil;
}

func ensureChannelHealth() error {
	if channel == nil || channel.IsClosed() {
		return establishChannel();
	}
	return nil;
}

func handleMatchRequests() {
	msgs, _ := channel.ConsumeWithContext(context.Background(), "match_req", "mmih", true, false, false, false, nil)
	for msg := range msgs {
		var request messages.MatchRequest;
		err := proto.Unmarshal(msg.Body, &request)
		if err != nil {
			log.Println("ERROR: Couldnt unmarshall proto")
		}
		MatchRequests<-request
	}
}
