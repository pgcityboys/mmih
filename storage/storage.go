package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mmih/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Redis consits of two types of keymaps:
// 1st one is the category map: it stores all the rooms of a given category, keeping info about users participating in the rooms
// 2nd one stores the latest added room id in each category

type RoomInfo struct {
	Users []string
	Category string
	MaxUsers int
	Description string
}

type CategoryInfo struct {
	Rooms []string
}

var mainClient redis.Client;
var metaClient redis.Client;
var ctx context.Context = context.Background();

func InitializeRedisConnection() {
	redisAddress := utils.EnvWithDefaults("REDIS_ADDRESS", "localhost:6379")
	mainClient = *redis.NewClient(&redis.Options{
		Addr: redisAddress,
		Password: "",
		DB: 1,
	})
	metaClient = *redis.NewClient(&redis.Options{
		Addr: redisAddress,
		Password: "",
		DB: 2,
	})
	log.Println("Connected to Redis client")
}

func GetRoom(id string) (RoomInfo, error) {
	val, err := mainClient.Get(context.Background(), id).Result()
	if err == redis.Nil {
		log.Printf("Room with ID %s does not exist", id)
		return RoomInfo{}, err;
	} else if err != nil {
		return RoomInfo{}, err;
	} else {
		var info RoomInfo;
		json.Unmarshal([]byte(val), &info)
		return info, nil
	}
}

func setRoom(id string, data *RoomInfo) {
	binaryData, _ := json.Marshal(data)
	mainClient.Set(ctx, id, binaryData, 24 * time.Hour);
}

func CreateEmptyRoom(data *RoomInfo) (id string) {
	id = uuid.New().String()
	binaryData, _ := json.Marshal(data)
	mainClient.Set(context.Background(), id, string(binaryData), 24 * time.Hour)
	return id;
}

// Returns ids of all rooms from a category
func CategoryRooms(category string) ([]string, error) {
	val, err := metaClient.Get(context.Background(), category).Result()
	if err == redis.Nil {
		return nil, err
	} else if err != nil {
		return nil, err
	} else {
		var info CategoryInfo;
		json.Unmarshal([]byte(val), &info)
		return info.Rooms, nil
	}
}

func JoinRoom(user, room string) error {
	info, err := GetRoom(room);
	if err != nil {
		return err
	}
	if len(info.Users) == info.MaxUsers {
		return fmt.Errorf("max users for room reached")
	}
	info.Users = append(info.Users, user)
	setRoom(room, &info);
	return nil
}
