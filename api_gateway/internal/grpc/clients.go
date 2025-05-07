package grpc

import (
	playlistpb "github.com/Zhan028/Music_Service/playlistService/proto"
	trackspb "github.com/Zhan028/Music_Service/track-service/proto"
	userpb "github.com/Zhan028/Music_Service/userService/proto"
	"google.golang.org/grpc"
)

type Clients struct {
	UserClient     userpb.UserServiceClient
	PlaylistClient playlistpb.PlaylistServiceClient
	TracksClient   trackspb.TrackServiceClient
}

func NewClients() *Clients {
	userConn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
	playlistConn, _ := grpc.Dial("localhost:50052", grpc.WithInsecure())
	trackConn, _ := grpc.Dial("localhost:50053", grpc.WithInsecure())
	return &Clients{
		UserClient:     userpb.NewUserServiceClient(userConn),
		PlaylistClient: playlistpb.NewPlaylistServiceClient(playlistConn),
		TracksClient:   trackspb.NewTrackServiceClient(trackConn),
	}

}
