package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type PostEvent struct {
	UserID    uint   `json:"user_id"`
	PostID    uint   `json:"post_id"`
	ImageURL  string `json:"image_url"`
	Timestamp string `json:"timestamp"`
}

type Follow struct {
	FollowerID uint `json:"follower_id"`
	FolloweeID uint `json:"followee_id"`
}

type GetFollowersResponse struct {
	Data []Follow `json:"data"`
}

type FeedGenerationService struct {
	kafkaConsumer *kafka.Consumer
	redisClient   *redis.Client
	followSvcURL  string
}

func NewFeedGenerationService() *FeedGenerationService {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "feed-generation",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("failed to create kafka consumer: %v", err)
	}

	err = consumer.Subscribe("post-created", nil)
	if err != nil {
		log.Fatalf("failed to subscribe to kafka topic: %v", err)
	}

	return &FeedGenerationService{
		kafkaConsumer: consumer,
		redisClient:   redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
		followSvcURL:  "http://localhost:8081",
	}
}

func (s *FeedGenerationService) processMessages() {
	for {
		msg, err := s.kafkaConsumer.ReadMessage(-1)
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		// Parse the message
		var postEvent PostEvent
		if err := json.Unmarshal(msg.Value, &postEvent); err != nil {
			log.Println("Failed to unmarshal post event:", err)
			continue
		}

		// Fetch followers from follow-svc
		followers, err := s.getFollowers(postEvent.UserID)
		if err != nil {
			log.Println("Failed to get followers:", err)
			continue
		}

		// Update feeds of all followers
		for _, follower := range followers {
			log.Printf("Updating feed for follower %d\n", follower.FollowerID)

			err := s.updateFeed(follower.FollowerID, postEvent)
			if err != nil {
				log.Println("Failed to update feed for follower:", follower.FollowerID)
			}
		}
	}
}

func (s *FeedGenerationService) getFollowers(userID uint) ([]Follow, error) {
	resp, err := http.Get(fmt.Sprintf("%s/followers/%d", s.followSvcURL, userID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var getFollowersResp GetFollowersResponse
	err = json.NewDecoder(resp.Body).Decode(&getFollowersResp)
	return getFollowersResp.Data, err
}

func (s *FeedGenerationService) updateFeed(followerID uint, postEvent PostEvent) error {
	timelineKey := fmt.Sprintf("timeline:%d", followerID)
	postEntry := fmt.Sprintf("%d|%s|%s", postEvent.PostID, postEvent.ImageURL, postEvent.Timestamp)
	err := s.redisClient.LPush(context.Background(), timelineKey, postEntry).Err()
	if err != nil {
		return err
	}
	return s.redisClient.LTrim(context.Background(), timelineKey, 0, 99).Err()
}

func main() {
	service := NewFeedGenerationService()
	service.processMessages()
}
