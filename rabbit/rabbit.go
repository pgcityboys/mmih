package rabbit

import (
	"fmt"
	"log"
	"mmih/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var connection *amqp.Connection;
var channel *amqp.Channel;

const TIMEOUT time.Duration = 5;
const RETRIES int = 5;
const SEND_TIMEOUT time.Duration = 10;


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

