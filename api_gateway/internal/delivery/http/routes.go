package http

import (
	"github.com/Zhan028/Music_Service/api_gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, h *Handler, jwtSecret string) {
	// Public routes
	router.POST("/register", h.RegisterUser) //robit
	router.POST("/login", h.Login)           //robit

	// Protected routes
	auth := router.Group("/")
	auth.Use(middleware.AuthMiddleware(jwtSecret))

	{
		//user CRUD
		auth.PUT("/userUpdate", h.UpdateUser)    //robit
		auth.DELETE("/userDelete", h.DeleteUser) //robit
		auth.GET("userProfile", h.GetUserByID)   //robit
		// Playlist CRUD
		auth.POST("/playlists", h.CreatePlaylist)       //robit
		auth.GET("/playlists", h.GetUserPlaylists)      //r
		auth.GET("/playlists/:id", h.GetPlaylist)       //r
		auth.DELETE("/playlists/:id", h.DeletePlaylist) //r
		auth.PUT("playlists/:id/put", h.AddTrackToPlaylist)
		// Track operations
		auth.POST("/playlists/:id/tracks", h.AddTrackToPlaylist)
		auth.DELETE("/playlists/:id/tracks/:trackId", h.RemoveTrackFromPlaylist)
		//Track service
		auth.POST("/tracksC", h.CreateTrack)       //r
		auth.DELETE("/tracksD", h.RemoveTrack)     //r
		auth.GET("/tracks/:id", h.GetTrack)        //r
		auth.PUT("/tracks/:id/put", h.UpdateTrack) //r

	}
}
