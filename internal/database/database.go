package database

import (
	"cloud.google.com/go/firestore"
	"context"
)

const (
	photosCollectionName = "photos"
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

func (d FirestoreDatabase) Update(ctx context.Context, id string, fields map[string]interface{}) error {
	cmd := createUpdateCommands(fields)
	_, err := d.client.Collection(photosCollectionName).Doc(id).Update(ctx, cmd)
	return err
}

func createUpdateCommands(fields map[string]interface{}) []firestore.Update {
	var cmd []firestore.Update
	for k, v := range fields {
		cmd = append(cmd, firestore.Update{
			Path:  k,
			Value: v,
		})
	}
	return cmd
}
