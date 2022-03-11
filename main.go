package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/config"
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"github.com/alancesar/photo-gallery/thumbs/domain/thumbs"
	"github.com/alancesar/photo-gallery/thumbs/internal/bucket"
	"github.com/alancesar/photo-gallery/thumbs/internal/listener"
	"github.com/alancesar/photo-gallery/thumbs/internal/publisher"
	"github.com/alancesar/photo-gallery/thumbs/usecase"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"os/signal"
)

const (
	configFileEnv        = "CONFIG_FILE"
	projectIDKey         = "PROJECT_ID"
	thumbsSubscriptionID = "thumbs"
	workerSubscriptionID = "worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	configFilePath := os.Getenv(configFileEnv)
	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	configs, err := config.Load(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	projectID := os.Getenv(projectIDKey)
	pubSubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalln(err)
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	handle := storageClient.Bucket(fmt.Sprintf("%s.appspot.com", projectID))
	photosBucket := bucket.New(handle)

	subscription := pubSubClient.Subscription(thumbsSubscriptionID)
	topic := pubSubClient.Topic(workerSubscriptionID)

	p := publisher.New[thumbs.Thumbs](topic)

	imageProcessor := image.NewProcessor()
	uc := usecase.NewThumbs(photosBucket, imageProcessor, p)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		l := listener.New[photo.Photo](subscription)
		if err := l.Listen(ctx, func(ctx context.Context, photo photo.Photo) error {
			log.Printf("received %s", photo.Filename)
			return uc.CreateThumbnails(ctx, photo.Filename, configs.Thumbs.Dimensions)
		}); err != nil {
			log.Println(err)
		}
	}()

	for {
		select {
		case <-signals:
			log.Println("shutting down...")
			cancel()
		case <-ctx.Done():
			log.Fatalln(ctx.Err())
		}
	}
}
