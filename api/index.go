package handler

import (
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Poll struct {
	ID       bson.ObjectID `json:"id"`
	Question string        `json:"question"`
	Options  []string      `json:"options"`
	Votes    []int         `json:"votes"`
}

var (
	polls []Poll
	mu    sync.Mutex
)

func handler() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Live Polling Tool API is running",
		})
	})

	r.POST("/api/polls", func(c *gin.Context) {
		var poll Poll

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
		poll.Votes = make([]int, len(poll.Options))

		mu.Lock()
		polls = append(polls, poll)
		mu.Unlock()

		c.JSON(http.StatusCreated, poll)
	})

	r.GET("/api/polls", func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		c.JSON(http.StatusOK, polls)
	})

	r.POST("/api/polls/:id/vote", func(c *gin.Context) {
		var vote struct {
			Option int `json:"option"`
		}

		if err := c.ShouldBindJSON(&vote); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote"})
			return
		}

		mu.Lock()
		defer mu.Unlock()

		for i := range polls {
			if polls[i].ID.Hex() == c.Param("id") {
				if vote.Option < 0 || vote.Option >= len(polls[i].Options) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option"})
					return
				}

				polls[i].Votes[vote.Option]++

				c.JSON(http.StatusOK, polls[i])
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
	})

	return r
}

func Handler(w http.ResponseWriter, req *http.Request) {
	handler().ServeHTTP(w, req)
}