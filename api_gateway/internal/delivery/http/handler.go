package http

import (
	"context"
	"github.com/Zhanbatyr06/Music_Service/api_gateway/internal/grpc"
	"net/http"

	playlistpb "github.com/Zhan028/Music_Service/playlistService/proto"
	userpb "github.com/Zhan028/Music_Service/userService/proto"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	clients *grpc.Clients
}

func NewHandler(clients *grpc.Clients) *Handler {
	return &Handler{clients: clients}
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var req userpb.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.clients.UserClient.RegisterUser(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) CreatePlaylist(c *gin.Context) {
	var req playlistpb.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.clients.PlaylistClient.CreatePlaylist(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
