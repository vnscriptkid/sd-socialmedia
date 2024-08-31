package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type FeedService struct {
	redisClient *redis.Client
}

var ctx = context.Background()

func NewFeedService() *FeedService {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return &FeedService{redisClient: rdb}
}

func (s *FeedService) UpdateFeed(followerID uint, postID uint, imageURL string) error {
	timelineKey := fmt.Sprintf("timeline:%d", followerID)
	postEntry := fmt.Sprintf("%d|%s", postID, imageURL)
	err := s.redisClient.LPush(ctx, timelineKey, postEntry).Err()
	if err != nil {
		return err
	}
	// Trim timeline to last 100 posts
	return s.redisClient.LTrim(ctx, timelineKey, 0, 99).Err()
}

func (s *FeedService) GetFeed(c *gin.Context) {
	userID := c.Param("user_id")
	timelineKey := fmt.Sprintf("timeline:%s", userID)
	feed, err := s.redisClient.LRange(ctx, timelineKey, 0, 99).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": feed})
}

func main() {
	feedService := NewFeedService()

	r := gin.Default()
	r.GET("/feeds/:user_id", feedService.GetFeed)
	r.Run(":8082")
}
