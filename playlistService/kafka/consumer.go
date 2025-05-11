package kafka

import (
	"context"
	"github.com/Zhan028/Music_Service/playlistService/internal/usecase"
	"github.com/segmentio/kafka-go"
	"log"
)

func StartTrackConsumer(uc usecase.PlaylistUseCase) {
	r := kafka.NewReader(kafka.ReaderConfig{
		// TO DO: перенести енвы(конфиг)
		Brokers: []string{"localhost:9092"},
		Topic:   "track.created",
		GroupID: "playlist-consumer-group",
	})

	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Printf("kafka read error: %v", err)
				continue
			}

			// TO DO: добавить способность читать несколько usecase
			err = uc.AddToNewPlaylist(context.Background(), m)
			if err != nil {
				log.Printf("could not add track to Новинки: %v", err)
			}
		}
	}()
}
