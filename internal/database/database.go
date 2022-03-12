package database

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/alancesar/photo-gallery/worker/domain/metadata"
	"github.com/alancesar/photo-gallery/worker/domain/thumb"
)

const (
	photosCollectionName = "photos"
	exifFieldName        = "exif"
	thumbsFieldName      = "thumbs"
)

type (
	FirestoreDatabase struct {
		client *firestore.Client
	}
)

func NewFirestoreDatabase(client *firestore.Client) *FirestoreDatabase {
	return &FirestoreDatabase{
		client: client,
	}
}

func (d FirestoreDatabase) InsertExif(ctx context.Context, id string, exif metadata.Exif) error {
	_, err := d.client.Collection(photosCollectionName).Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  exifFieldName,
			Value: exif,
		},
	})

	return err
}

func (d FirestoreDatabase) InsertThumbnails(ctx context.Context, id string, thumbs []thumb.Thumbnail) error {
	_, err := d.client.
		Collection(photosCollectionName).
		Doc(id).
		Update(ctx, []firestore.Update{
			{
				Path:  thumbsFieldName,
				Value: thumbs,
			},
		})

	return err
}
