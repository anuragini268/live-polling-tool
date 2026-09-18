package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"live-polling-tool/backend/config"
	"live-polling-tool/backend/models"
)

const pollsKey = "live_polling_polls"

func loadPolls() ([]models.Poll, error) {
	data, err := config.RedisClient.Get(context.Background(), pollsKey).Result()
	if err != nil {
		return []models.Poll{}, nil
	}

	var polls []models.Poll

	if err := json.Unmarshal([]byte(data), &polls); err != nil {
		return nil, err
	}

	return polls, nil
}

func savePolls(polls []models.Poll) error {
	data, err := json.Marshal(polls)
	if err != nil {
		return err
	}

	return config.RedisClient.Set(
		context.Background(),
		pollsKey,
		data,
		0,
	).Err()
}

func CreatePoll(c *gin.Context) {
	var poll models.Poll

	if err := c.ShouldBindJSON(&poll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if poll.Question == "" || len(poll.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question and at least 2 options are required",
		})
		return
	}

	poll.ID = bson.NewObjectID()
	poll.Votes = make([]int, len(poll.Options))

	polls, err := loadPolls()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load polls",
		})
		return
	}

	polls = append(polls, poll)

	if err := savePolls(polls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save poll",
		})
		return
	}

	c.JSON(http.StatusCreated, poll)
}

func GetPolls(c *gin.Context) {
	polls, err := loadPolls()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load polls",
		})
		return
	}

	c.JSON(http.StatusOK, polls)
}

func VotePoll(c *gin.Context) {
	pollID := c.Param("id")

	var optionIndex struct {
		Option int `json:"option"`
	}

	if err := c.ShouldBindJSON(&optionIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid vote",
		})
		return
	}

	polls, err := loadPolls()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load polls",
		})
		return
	}

	for i := range polls {
		if polls[i].ID.Hex() == pollID {

			if optionIndex.Option < 0 ||
				optionIndex.Option >= len(polls[i].Options) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid option",
				})
				return
			}

			polls[i].Votes[optionIndex.Option]++

			if err := savePolls(polls); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to save vote",
				})
				return
			}

			c.JSON(http.StatusOK, polls[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Poll not found",
	})
}