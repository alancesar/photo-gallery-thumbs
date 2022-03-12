package main

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/worker/config"
	"github.com/alancesar/photo-gallery/worker/domain/photo"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
	"github.com/alancesar/photo-gallery/worker/internal/bucket"
	"github.com/alancesar/photo-gallery/worker/internal/database"
	"github.com/alancesar/photo-gallery/worker/internal/extractor"
	"github.com/alancesar/photo-gallery/worker/internal/listener"
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

	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalln(err)
	}

	handle := storageClient.Bucket(fmt.Sprintf("%s.appspot.com", projectID))
	photosBucket := bucket.New(handle)

	imageProcessor := thumb.NewProcessor()
	thumbnailsUseCase := usecase.NewThumbnails(photosBucket, imageProcessor)
	exifUseCase := usecase.NewExif(photosBucket, extractor.Exif)

	subscription := pubSubClient.Subscription(thumbsSubscriptionID)
	l := listener.New[photo.Photo](subscription)

	db := database.NewFirestoreDatabase(firestoreClient)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		if err := l.Listen(ctx, func(ctx context.Context, incoming photo.Photo) error {
			log.Printf("received %s", incoming.Filename)
			group, _ := errgroup.WithContext(ctx)
			group.Go(func() error {
				thumbs, err := thumbnailsUseCase.Execute(ctx, incoming.Filename, configs.Thumbs.Dimensions)
				if err != nil {
					return err
				}

				return db.InsertThumbnails(ctx, incoming.ID, thumbs)
			})

			group.Go(func() error {
				exif, err := exifUseCase.Execute(ctx, incoming.Filename)
				if err != nil {
					return err
				}

				return db.InsertExif(ctx, incoming.ID, exif)
			})

			return group.Wait()
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
