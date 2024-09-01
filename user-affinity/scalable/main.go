package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID uint `gorm:"primaryKey"`
}

type FollowState string

const (
	FollowStateIsFollowedBy FollowState = "IS_FOLLOWED_BY"
	FollowStateFollow       FollowState = "FOLLOW"
)

type Follow struct {
	Left      uint        `gorm:"primaryKey"`
	Right     uint        `gorm:"index:idx_left_right_state,unique"`
	State     FollowState `gorm:"primaryKey;index:idx_left_right_state,unique"`
	CreatedAt time.Time   `gorm:"primaryKey"`
}

var db *gorm.DB

func main() {
	var err error
	dsn := "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto Migrate the schema
	db.AutoMigrate(&User{}, &Follow{})

	r := gin.Default()

	r.GET("/followers/:username", getFollowers)
	r.GET("/following/:username", getFollowing)
	r.GET("/is-following/:follower/:followed", isFollowing)
	r.POST("/follow", followUser)

	r.Run(":8080")
}

func getFollowers(c *gin.Context) {
	userID := c.Param("user_id")
	var user User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followerIDs []uint

	// (:Left) is followed by right
	db.Model(&Follow{}).
		Where("left = ? AND state = ?", user.ID, FollowStateIsFollowedBy).
		Pluck("right", &followerIDs)

	c.JSON(http.StatusOK, gin.H{"follower_ids": followerIDs})
}

func getFollowing(c *gin.Context) {
	userID := c.Param("user_id")
	var user User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followingIDs []uint
	// (:Left) follows right
	db.Model(&Follow{}).
		Where("left = ? AND state = ?", user.ID, FollowStateFollow).
		Pluck("right", &followingIDs)

	c.JSON(http.StatusOK, gin.H{"following_ids": followingIDs})
}

func isFollowing(c *gin.Context) {
	followerUserID := c.Param("follower")
	followedUserID := c.Param("followed")

	var follower, followed User
	if err := db.Where("id = ?", followerUserID).First(&follower).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Follower not found"})
		return
	}
	if err := db.Where("id = ?", followedUserID).First(&followed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Followed user not found"})
		return
	}

	var follow Follow
	result := db.Where("left = ? AND right = ? AND state = ?", follower.ID, followed.ID, FollowStateFollow).First(&follow)

	isFollowing := result.Error == nil

	c.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}

func followUser(c *gin.Context) {
	left, _ := strconv.ParseUint(c.Query("source"), 10, 32)
	right, _ := strconv.ParseUint(c.Query("dest"), 10, 32)

	createdAt := time.Now()

	follow1 := Follow{
		Left:      uint(left),
		Right:     uint(right),
		State:     FollowStateFollow,
		CreatedAt: createdAt,
	}

	follow2 := Follow{
		Left:      uint(left),
		Right:     uint(right),
		State:     FollowStateIsFollowedBy,
		CreatedAt: createdAt,
	}

	err := db.CreateInBatches([]Follow{follow1, follow2}, 2).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Follow relationship created"})
}

func init() {
	// Initialize database connection
	var err error
	dsn := "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// Auto Migrate the schema
	err = db.AutoMigrate(&User{}, &Follow{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}
