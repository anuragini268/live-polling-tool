package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"live-polling-tool/backend/config"
	"live-polling-tool/backend/models"
)

func generateShareCode() string {
	return bson.NewObjectID().Hex()[:10]
}

// CreatePoll creates a new poll for the authenticated user.
func CreatePoll(c *gin.Context) {
	var req struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	req.Question = strings.TrimSpace(req.Question)

	validOptions := make([]string, 0, len(req.Options))

	for _, option := range req.Options {
		option = strings.TrimSpace(option)

		if option != "" {
			validOptions = append(validOptions, option)
		}
	}

	if req.Question == "" || len(validOptions) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Question and at least 2 options are required",
		})
		return
	}

	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	ownerID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	ownerIDString, ok := ownerID.(string)

	if !ok || ownerIDString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user",
		})
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	poll := models.Poll{
		ID:        bson.NewObjectID(),
		OwnerID:   ownerIDString,
		ShareCode: generateShareCode(),
		Question:  req.Question,
		Options:   validOptions,
		Votes:     make([]int, len(validOptions)),
		CreatedAt: time.Now(),
	}

	_, err := config.MongoDB.Collection("polls").InsertOne(
		ctx,
		poll,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create poll",
		})
		return
	}

	c.JSON(http.StatusCreated, poll)
}

// GetPolls returns all polls.
func GetPolls(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	cursor, err := config.MongoDB.Collection("polls").Find(
		ctx,
		bson.M{},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load polls",
		})
		return
	}

	defer cursor.Close(ctx)

	var polls []models.Poll

	if err := cursor.All(ctx, &polls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read polls",
		})
		return
	}

	if polls == nil {
		polls = []models.Poll{}
	}

	c.JSON(http.StatusOK, polls)
}

// GetPoll returns a single poll by ID or share code.
func GetPoll(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	value := c.Param("id")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	var poll models.Poll

	objectID, err := bson.ObjectIDFromHex(value)

	if err == nil {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"_id": objectID},
		).Decode(&poll)
	} else {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"share_code": value},
		).Decode(&poll)
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load poll",
		})
		return
	}

	c.JSON(http.StatusOK, poll)
}

// VotePoll records one vote per voter.
func VotePoll(c *gin.Context) {
	if config.MongoDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database is not connected",
		})
		return
	}

	value := c.Param("id")

	var req struct {
		Option  int    `json:"option"`
		VoterID string `json:"voter_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid vote",
		})
		return
	}

	req.VoterID = strings.TrimSpace(req.VoterID)

	if req.VoterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Voter ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	var poll models.Poll
	var err error

	objectID, objectErr := bson.ObjectIDFromHex(value)

	if objectErr == nil {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"_id": objectID},
		).Decode(&poll)
	} else {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"share_code": value},
		).Decode(&poll)
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load poll",
		})
		return
	}

	if req.Option < 0 || req.Option >= len(poll.Options) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid option",
		})
		return
	}

	votesCollection := config.MongoDB.Collection("votes")

	vote := models.Vote{
		ID:        bson.NewObjectID(),
		PollID:    poll.ID,
		VoterID:   req.VoterID,
		Option:    req.Option,
		CreatedAt: time.Now(),
	}

	// Insert vote record.
	// A unique MongoDB index on poll_id + voter_id
	// will prevent the same voter from voting twice.
	_, err = votesCollection.InsertOne(ctx, vote)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You have already voted in this poll",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record vote",
		})
		return
	}

	// Correctly update the selected option.
	voteField := "votes." + strconv.Itoa(req.Option)

	_, err = config.MongoDB.Collection("polls").UpdateOne(
		ctx,
		bson.M{"_id": poll.ID},
		bson.M{
			"$inc": bson.M{
				voteField: 1,
			},
		},
	)

	if err != nil {
		// Roll back vote record if poll update fails.
		_, _ = votesCollection.DeleteOne(
			ctx,
			bson.M{"_id": vote.ID},
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update vote count",
		})
		return
	}

	// Load updated poll.
	err = config.MongoDB.Collection("polls").FindOne(
		ctx,
		bson.M{"_id": poll.ID},
	).Decode(&poll)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated poll",
		})
		return
	}

	// Publish updated poll through Redis.
	if config.RedisClient != nil {
		data, marshalErr := json.Marshal(poll)

		if marshalErr == nil {
			channel := "poll:" + poll.ID.Hex() + ":updates"

			err = config.RedisClient.Publish(
				ctx,
				channel,
				data,
			).Err()

			if err != nil {
				// Vote is already saved successfully.
				// Log the error but don't undo the vote.
				// The next vote/update can publish a fresh state.
				println("Redis publish failed:", err.Error())
			}
		}
	}

	c.JSON(http.StatusOK, poll)
}

// StreamPoll sends live poll updates using Redis Pub/Sub.
func StreamPoll(c *gin.Context) {
	if config.MongoDB == nil || config.RedisClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Real-time service is not connected",
		})
		return
	}

	value := c.Param("id")

	ctx := c.Request.Context()

	var poll models.Poll
	var err error

	objectID, objectErr := bson.ObjectIDFromHex(value)

	if objectErr == nil {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"_id": objectID},
		).Decode(&poll)
	} else {
		err = config.MongoDB.Collection("polls").FindOne(
			ctx,
			bson.M{"share_code": value},
		).Decode(&poll)
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Poll not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load poll",
		})
		return
	}

	channelName := "poll:" + poll.ID.Hex() + ":updates"

	pubsub := config.RedisClient.Subscribe(
		ctx,
		channelName,
	)

	defer pubsub.Close()

	// Wait until Redis confirms the subscription.
	_, err = pubsub.Receive(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to real-time channel",
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Send current poll immediately.
	initialData, err := json.Marshal(poll)

	if err == nil {
		c.SSEvent(
			"poll",
			string(initialData),
		)

		c.Writer.Flush()
	}

	// Listen for Redis Pub/Sub updates.
	channel := pubsub.Channel()

	for {
		select {

		case <-ctx.Done():
			return

		case message, ok := <-channel:

			if !ok {
				return
			}

			c.SSEvent(
				"poll",
				message.Payload,
			)

			c.Writer.Flush()
		}
	}
}
