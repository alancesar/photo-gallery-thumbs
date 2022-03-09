package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/alancesar/photo-gallery/thumbs/config"
	"github.com/alancesar/photo-gallery/thumbs/domain/image"
	"github.com/alancesar/photo-gallery/thumbs/domain/photo"
	"github.com/alancesar/photo-gallery/thumbs/domain/thumbs"
	"github.com/alancesar/photo-gallery/thumbs/internal/listener"
	"github.com/alancesar/photo-gallery/thumbs/internal/publisher"
	"github.com/alancesar/photo-gallery/thumbs/internal/storage"
	"github.com/alancesar/photo-gallery/thumbs/usecase"
	_ "github.com/joho/godotenv/autoload"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
	"os/signal"
)

const (
	minioEndpointEnv     = "MINIO_ENDPOINT"
	minioRootUserEnv     = "MINIO_ROOT_USER"
	minioRootPasswordEnv = "MINIO_ROOT_PASSWORD"
	photosBucketEnv      = "PHOTOS_BUCKET"
	thumbsBucketEnv      = "THUMBS_BUCKET"
	configFileEnv        = "CONFIG_FILE"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	configFile := os.Getenv(configFileEnv)
	configs, err := config.Load(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := minio.New(os.Getenv(minioEndpointEnv), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv(minioRootUserEnv), os.Getenv(minioRootPasswordEnv), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	pubSubClient, err := pubsub.NewClient(ctx, "photo-gallery")
	if err != nil {
		log.Fatalln(err)
	}

	subscription := pubSubClient.Subscription("worker")
	topic := pubSubClient.Topic("photos")

	photoStorage := storage.NewMinioStorage(client, os.Getenv(photosBucketEnv))
	thumbsStorage := storage.NewMinioStorage(client, os.Getenv(thumbsBucketEnv))
	imageProcessor := image.NewProcessor()
	thumbsPublisher := publisher.New[thumbs.Thumbs](topic)

	w := usecase.NewThumbs(photoStorage, thumbsStorage, imageProcessor, thumbsPublisher)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		l := listener.New[photo.Photo](subscription)
		if err := l.Listen(ctx, func(ctx context.Context, photo photo.Photo) error {
			return w.CreateThumbnails(ctx, usecase.CreateThumbsRequest{
				Filename:   photo.Filename,
				Metadata:   photo.Metadata,
				Dimensions: configs.Thumbs.Dimensions,
			})
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
