package routes

import (
	"github.com/gin-gonic/gin"
	"live-polling-tool/backend/controllers"
)

func PollRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/polls", controllers.CreatePoll)
		api.GET("/polls", controllers.GetPolls)
		api.POST("/polls/:id/vote", controllers.VotePoll)
	}
}