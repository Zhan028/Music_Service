package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.POST("/register", h.RegisterUser)
	r.POST("/playlists", h.CreatePlaylist)
}
