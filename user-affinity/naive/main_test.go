package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{}, &Follow{})
	return db
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/followers/:username", getFollowers)
	r.GET("/following/:username", getFollowing)
	r.GET("/is-following/:follower/:followed", isFollowing)
	return r
}

func TestGetFollowers(t *testing.T) {
	db = setupTestDB()
	router := setupRouter()

	// Create test users
	db.Create(&User{ID: 1, Username: "alice"})
	db.Create(&User{ID: 2, Username: "bob"})
	db.Create(&Follow{FollowerID: 2, FollowedID: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/followers/alice", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string][]uint
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, []uint{2}, response["follower_ids"])
}

func TestGetFollowing(t *testing.T) {
	db = setupTestDB()
	router := setupRouter()

	// Create test users
	db.Create(&User{ID: 1, Username: "alice"})
	db.Create(&User{ID: 2, Username: "bob"})
	db.Create(&Follow{FollowerID: 1, FollowedID: 2})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/following/alice", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string][]uint
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, []uint{2}, response["following_ids"])
}

func TestIsFollowing(t *testing.T) {
	db = setupTestDB()
	router := setupRouter()

	// Create test users
	db.Create(&User{ID: 1, Username: "alice"})
	db.Create(&User{ID: 2, Username: "bob"})
	db.Create(&Follow{FollowerID: 1, FollowedID: 2})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/is-following/alice/bob", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]bool
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["is_following"])
}

func TestUserNotFound(t *testing.T) {
	db = setupTestDB()
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/followers/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "User not found", response["error"])
}
