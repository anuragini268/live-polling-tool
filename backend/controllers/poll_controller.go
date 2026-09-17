package controllers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"live-polling-tool/backend/models"
)

var (
	polls []models.Poll
	mu    sync.Mutex
)

func CreatePoll(c *gin.Context) {
	var poll models.Poll

	if err := c.ShouldBindJSON(&poll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if poll.Question == "" || len(poll.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question and at least 2 options are required",
		})
		return
	}

	poll.ID = bson.NewObjectID()

	// Initialize vote count for each option
	poll.Votes = make([]int, len(poll.Options))

	mu.Lock()
	polls = append(polls, poll)
	mu.Unlock()

	c.JSON(http.StatusCreated, poll)
}

func GetPolls(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, polls)
}

func VotePoll(c *gin.Context) {
	pollID := c.Param("id")

	optionIndex := struct {
		Option int `json:"option"`
	}{}

	if err := c.ShouldBindJSON(&optionIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid vote",
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

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

			c.JSON(http.StatusOK, polls[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Poll not found",
	})
}