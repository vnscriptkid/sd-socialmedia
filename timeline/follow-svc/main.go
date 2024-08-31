package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Follow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"not null" json:"follower_id"`
	FolloweeID uint      `gorm:"not null" json:"followee_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CreateFollowInput struct {
	FollowerID uint `json:"follower_id" binding:"required"`
	FolloweeID uint `json:"followee_id" binding:"required"`
}

var DB *gorm.DB

func initDB() {
	var err error
	dsn := "host=localhost user=postgres password=123456 dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	DB.AutoMigrate(&Follow{})
}

func createFollow(c *gin.Context) {
	var input CreateFollowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	follow := Follow{FollowerID: input.FollowerID, FolloweeID: input.FolloweeID}
	err := DB.Create(&follow).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": follow})
}

func getFollowers(c *gin.Context) {
	var follows []Follow
	followeeID := c.Param("followee_id")
	err := DB.Where("followee_id = ?", followeeID).Find(&follows).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": follows})
}

func main() {
	initDB()

	r := gin.Default()
	r.GET("/followers/:followee_id", getFollowers)
	r.POST("/follows", createFollow)
	r.Run(":8081")
}
