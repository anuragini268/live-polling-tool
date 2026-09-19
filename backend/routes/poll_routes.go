package routes

import (
	"github.com/gin-gonic/gin"

	"live-polling-tool/backend/controllers"
	"live-polling-tool/backend/middleware"
)

func PollRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Authentication
		api.POST("/auth/signup", controllers.Signup)
		api.POST("/auth/login", controllers.Login)

		// Poll creation - authentication required
		api.POST(
			"/polls",
			middleware.AuthRequired(),
			controllers.CreatePoll,
		)

		// Public poll endpoints
		api.GET("/polls", controllers.GetPolls)
		api.GET("/polls/:id", controllers.GetPoll)

		// Voting
		api.POST("/polls/:id/vote", controllers.VotePoll)

		// Real-time live updates using Redis + SSE
		api.GET(
			"/polls/:id/stream",
			controllers.StreamPoll,
		)
	}
}
