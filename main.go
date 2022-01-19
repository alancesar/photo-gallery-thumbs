package main

import (
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/thumbs/api"
	"github.com/alancesar/photo-gallery/thumbs/config"
	"github.com/alancesar/photo-gallery/thumbs/consumer"
	"github.com/alancesar/photo-gallery/thumbs/database"
	"github.com/alancesar/photo-gallery/thumbs/image"
	"github.com/alancesar/photo-gallery/thumbs/pubsub"
	"github.com/alancesar/photo-gallery/thumbs/storage"
	"github.com/alancesar/photo-gallery/thumbs/worker"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	dbHostEnv            = "DB_HOST"
	dbUserEnv            = "DB_USER"
	dbPasswordEnv        = "DB_PASSWORD"
	dbNameEnv            = "DB_NAME"
	dbPortEnv            = "DB_PORT"
	minioEndpointEnv     = "MINIO_ENDPOINT"
	minioRootUserEnv     = "MINIO_ROOT_USER"
	minioRootPasswordEnv = "MINIO_ROOT_PASSWORD"
	photosBucketEnv      = "PHOTOS_BUCKET"
	thumbsBucketEnv      = "THUMBS_BUCKET"
	rabbitMQUrlEnv       = "RABBITMQ_URL"
	queueNameEnv         = "QUEUE_NAME"
	exchangeNameEnv      = "EXCHANGE_NAME"
	configFileEnv        = "CONFIG_FILE"
)

func main() {
	configFile := os.Getenv(configFileEnv)
	configs, err := config.Load(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", os.Getenv(dbHostEnv),
		os.Getenv(dbUserEnv), os.Getenv(dbPasswordEnv), os.Getenv(dbNameEnv), os.Getenv(dbPortEnv))
	dbConnection, err := database.NewConnection(dsn)
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

	connection, err := amqp.Dial(os.Getenv(rabbitMQUrlEnv))
	if err != nil {
		log.Fatalln(err)
	}
	defer func(conn *amqp.Connection) {
		_ = conn.Close()
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	defer func(conn *amqp.Channel) {
		_ = conn.Close()
	}(channel)

	db := image.NewDatabase(dbConnection)
	dimensions := configs.Thumbs.Dimensions
	publisher := pubsub.NewPublisher(channel, os.Getenv(exchangeNameEnv))
	bundle := worker.Bundle{
		PhotoStorage: storage.NewMinioStorage(client, os.Getenv(photosBucketEnv)),
		ThumbStorage: storage.NewMinioStorage(client, os.Getenv(thumbsBucketEnv)),
		Database:     db,
		Processor:    image.NewImagingProcessor(),
		Producer:     image.NewProducer(publisher),
		Dimensions:   dimensions,
	}
	w := worker.NewThumbsWorker(bundle)
	c := consumer.NewConsumer(w)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		subscriber := pubsub.NewSubscriber(channel)
		if err := subscriber.Subscribe(ctx, os.Getenv(queueNameEnv), c); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		engine := gin.Default()
		engine.Use(cors.Default())
		engine.Handle(http.MethodGet, "/api/thumbs/:filename", api.GetThumbsHandler(db))
		if err := engine.Run(":8082"); err != nil {
			log.Fatalln(err)
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
