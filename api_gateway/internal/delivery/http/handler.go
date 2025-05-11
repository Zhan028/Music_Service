package http

import (
	"context"
	"github.com/Zhan028/Music_Service/api_gateway/internal/grpc"
	"net/http"

	playlistpb "github.com/Zhan028/Music_Service/playlistService/proto"
	tracskpb "github.com/Zhan028/Music_Service/track-service/proto"
	userpb "github.com/Zhan028/Music_Service/userService/proto"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	clients *grpc.Clients
}

func NewHandler(clients *grpc.Clients) *Handler {
	return &Handler{clients: clients}
}

// ======= USER HANDLER =======

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

// ======= PLAYLIST HANDLER =======

type TrackInput struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Duration int32  `json:"duration"`
	Album    string `json:"album"`
}

type CreatePlaylistInput struct {
	Name        string       `json:"name"`
	UserID      string       `json:"user_id"`
	Description string       `json:"description"`
	Tracks      []TrackInput `json:"tracks"`
}

func (h *Handler) CreatePlaylist(c *gin.Context) {
	var input CreatePlaylistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var protoTracks []*playlistpb.Track
	for _, t := range input.Tracks {
		protoTracks = append(protoTracks, &playlistpb.Track{
			Title:    t.Title,
			Artist:   t.Artist,
			Duration: t.Duration,
			Album:    t.Album,
		})
	}

	req := &playlistpb.CreatePlaylistRequest{
		Name:        input.Name,
		UserId:      input.UserID,
		Description: input.Description,
		Tracks:      protoTracks,
	}

	resp, err := h.clients.PlaylistClient.CreatePlaylist(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetUserPlaylists(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query param is required"})
		return
	}

	req := &playlistpb.GetUserPlaylistsRequest{
		UserId: userID,
	}

	resp, err := h.clients.PlaylistClient.GetUserPlaylists(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ======= TRACK HANDLER =======
func (h *Handler) DeleteTrack(c *gin.Context) {
	var input tracskpb.DeleteTrackRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.clients.TracksClient.DeleteTrack(context.Background(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, resp)

}
func (h *Handler) CreateTrack(c *gin.Context) {
	var input tracskpb.CreateTrackRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.clients.TracksClient.CreateTrack(context.Background(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
