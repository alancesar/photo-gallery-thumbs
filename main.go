package main

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/worker/config"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/photo"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
	"github.com/alancesar/photo-gallery/worker/internal/bucket"
	"github.com/alancesar/photo-gallery/worker/internal/extractor"
	"github.com/alancesar/photo-gallery/worker/internal/listener"
	"github.com/alancesar/photo-gallery/worker/internal/publisher"
	"github.com/alancesar/photo-gallery/worker/presenter/message"
	"github.com/alancesar/photo-gallery/worker/usecase"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
)

const (
	configFileEnv        = "CONFIG_FILE"
	projectIDKey         = "PROJECT_ID"
	thumbsSubscriptionID = "thumbs"
	thumbsTopicID        = "thumbs"
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

	imageProcessor := thumb.NewProcessor()
	thumbnailsUseCase := usecase.NewThumbnails(photosBucket, imageProcessor)
	exifUseCase := usecase.NewExif(photosBucket, extractor.Exif)

	topic := pubSubClient.Topic(thumbsTopicID)
	p := publisher.New[message.Photo](topic)

	subscription := pubSubClient.Subscription(thumbsSubscriptionID)
	l := listener.New[photo.Photo](subscription)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		if err := l.Listen(ctx, func(ctx context.Context, incoming photo.Photo) error {
			log.Printf("received %s", incoming.Filename)
			var (
				exif   metadata.Exif
				thumbs []thumb.Thumbnail
			)

			group, _ := errgroup.WithContext(ctx)
			group.Go(func() error {
				var err error
				thumbs, err = thumbnailsUseCase.Execute(ctx, incoming.Filename, configs.Thumbs.Dimensions)
				return err
			})

			group.Go(func() error {
				var err error
				exif, err = exifUseCase.Execute(ctx, incoming.Filename)
				return err
			})

			if err := group.Wait(); err != nil {
				return err
			}

			p.Publish(ctx, message.Photo{
				ID:     incoming.ID,
				Thumbs: thumbs,
				Exif:   exif,
			})

			return nil
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
