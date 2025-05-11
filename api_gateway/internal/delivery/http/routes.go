package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.POST("/register", h.RegisterUser)
	r.POST("/playlists", h.CreatePlaylist)
	r.GET("/playlists", h.GetUserPlaylists)
	r.POST("/delete", h.DeleteTrack)
	r.POST("/createTrack", h.CreateTrack)
}
