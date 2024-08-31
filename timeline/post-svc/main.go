package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	ImageURL  string    `gorm:"not null" json:"image_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CreatePostInput struct {
	UserID   uint   `json:"user_id" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
}

var DB *gorm.DB
var kafkaProducer *kafka.Producer

var (
	topicPostCreated = "post-created"
)

func initDB() {
	var err error
	dsn := "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(&Post{})
}

func initKafka() {
	var err error
	kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		panic("failed to create kafka producer")
	}
}

func createPost(c *gin.Context) {
	var input CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post := Post{UserID: input.UserID, ImageURL: input.ImageURL}
	err := DB.Create(&post).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish event to Kafka
	event := map[string]interface{}{
		"user_id":   post.UserID,
		"post_id":   post.ID,
		"image_url": post.ImageURL,
		"timestamp": post.CreatedAt,
	}
	eventBytes, _ := json.Marshal(event)

	kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicPostCreated, Partition: kafka.PartitionAny},
		Value:          eventBytes,
	}, nil)

	c.JSON(http.StatusOK, gin.H{"data": post})
}

func main() {
	initDB()
	initKafka()

	r := gin.Default()
	r.POST("/posts", createPost)
	r.Run(":8080") // listen and serve on port 8080
}
