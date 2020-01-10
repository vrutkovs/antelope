package gcs

import (
	"context"
	"log"

	"cloud.google.com/go/storage"
)

const (
	bucketName = "origin-ci-test"
	basePrefix = "logs/"
)

func GetGCSBucket() *storage.BucketHandle {

	ctx := context.Background()

	// Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return client.Bucket(bucketName)
}
