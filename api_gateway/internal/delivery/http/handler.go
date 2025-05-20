package http

import (
	"context"
	"github.com/Zhan028/Music_Service/api_gateway/internal/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"

	playlistpb "github.com/Zhan028/Music_Service/playlistService/proto"
	trackspb "github.com/Zhan028/Music_Service/track-service/proto"
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

func (h *Handler) Login(c *gin.Context) {
	var req userpb.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.clients.UserClient.AuthenticateUser(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      resp.Token,
		"user_id":    resp.UserId,
		"expires_at": resp.ExpiresAt,
	})
}
func (h *Handler) UpdateUser(c *gin.Context) {
	var req userpb.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
	}

	resp, err := h.clients.UserClient.UpdateUserProfile(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}
func (h *Handler) DeleteUser(c *gin.Context) {
	var req userpb.UserID
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
	}

	resp, err := h.clients.UserClient.DeleteUser(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
	}
	c.JSON(http.StatusCreated, resp)
}
func (h *Handler) GetUserByID(c *gin.Context) {
	var req userpb.UserID
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
	}
	resp, err := h.clients.UserClient.GetUserProfile(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
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
	Description string       `json:"description"`
	Tracks      []TrackInput `json:"tracks"`
}

type PlaylistInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tracks      []TrackInput `json:"tracks"`
	IsPublic    bool         `json:"is_public"`
}

type UpdatePlaylistInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// CreatePlaylist создает новый плейлист
func (h *Handler) CreatePlaylist(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

	var input PlaylistInput
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
		UserId:      userID,
		Description: input.Description,
		Tracks:      protoTracks,
		IsPublic:    input.IsPublic,
	}

	resp, err := h.clients.PlaylistClient.CreatePlaylist(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetPlaylist возвращает плейлист по ID
func (h *Handler) GetPlaylist(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	playlistID := c.Param("id")

	req := &playlistpb.GetPlaylistRequest{
		Id:     playlistID,
		UserId: userID,
	}

	resp, err := h.clients.PlaylistClient.GetPlaylist(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserPlaylists возвращает все плейлисты пользователя
func (h *Handler) GetUserPlaylists(c *gin.Context) {
	userID := c.MustGet("user_id").(string)

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

// DeletePlaylist удаляет плейлист
func (h *Handler) DeletePlaylist(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	playlistID := c.Param("id") // Важно: используем Param, а не Query

	log.Printf("Delete request - PlaylistID: %s, UserID: %s", playlistID, userID)

	if playlistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "playlist ID is required"})
		return
	}

	req := &playlistpb.DeletePlaylistRequest{
		Id:     playlistID,
		UserId: userID,
	}

	resp, err := h.clients.PlaylistClient.DeletePlaylist(context.Background(), req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddTrackToPlaylist добавляет трек в плейлист
func (h *Handler) AddTrackToPlaylist(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	playlistID := c.Param("id")

	var input TrackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &playlistpb.AddTrackRequest{
		PlaylistId: playlistID,
		UserId:     userID,
		Track: &playlistpb.Track{
			Title:    input.Title,
			Artist:   input.Artist,
			Duration: input.Duration,
			Album:    input.Album,
		},
	}

	resp, err := h.clients.PlaylistClient.AddTrackToPlaylist(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// RemoveTrackFromPlaylist удаляет трек из плейлиста
func (h *Handler) RemoveTrackFromPlaylist(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	playlistID := c.Param("id")
	trackID := c.Param("trackId")

	req := &playlistpb.RemoveTrackRequest{
		PlaylistId: playlistID,
		TrackId:    trackID,
		UserId:     userID,
	}

	resp, err := h.clients.PlaylistClient.RemoveTrackFromPlaylist(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Track handler
func (h *Handler) CreateTrack(c *gin.Context) {
	var req trackspb.CreateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	resp, err := h.clients.TracksClient.CreateTrack(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, resp)

}
func (h *Handler) RemoveTrack(c *gin.Context) {
	var req trackspb.DeleteTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	resp, err := h.clients.TracksClient.DeleteTrack(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, resp)

}
func (h *Handler) GetTrack(c *gin.Context) {
	var req trackspb.GetTrackByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	resp, err := h.clients.TracksClient.GetTrackByID(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, resp)

}
func (h *Handler) UpdateTrack(c *gin.Context) {
	var req trackspb.UpdateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	resp, err := h.clients.TracksClient.UpdateTrack(context.Background(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, resp)

}
