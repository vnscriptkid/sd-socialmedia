package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
}

type Follow struct {
	ID         uint `gorm:"primaryKey"`
	FollowerID uint
	FollowedID uint
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

	r.Run(":8080")
}

func getFollowers(c *gin.Context) {
	username := c.Param("username")
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followerIDs []uint
	db.Model(&Follow{}).
		Where("followed_id = ?", user.ID).
		Pluck("follower_id", &followerIDs)

	c.JSON(http.StatusOK, gin.H{"follower_ids": followerIDs})
}

func getFollowing(c *gin.Context) {
	username := c.Param("username")
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followingIDs []uint
	db.Model(&Follow{}).
		Where("follower_id = ?", user.ID).
		Pluck("followed_id", &followingIDs)

	c.JSON(http.StatusOK, gin.H{"following_ids": followingIDs})
}

func isFollowing(c *gin.Context) {
	followerUsername := c.Param("follower")
	followedUsername := c.Param("followed")

	var follower, followed User
	if err := db.Where("username = ?", followerUsername).First(&follower).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Follower not found"})
		return
	}
	if err := db.Where("username = ?", followedUsername).First(&followed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Followed user not found"})
		return
	}

	var follow Follow
	result := db.Where("follower_id = ? AND followed_id = ?", follower.ID, followed.ID).First(&follow)

	isFollowing := result.Error == nil

	c.JSON(http.StatusOK, gin.H{"is_following": isFollowing})
}
