package main

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type (
	Storage interface {
		GetBlob(c context.Context, name string) ([]byte, error)
	}

	googleStorage struct {
		client *storage.Client
		bucket string
	}
)

func NewGoogleStorage(bucket string) (*googleStorage, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &googleStorage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *googleStorage) GetBlob(c context.Context, name string) ([]byte, error) {
	logger.Debugf("get blob %s/%s", s.bucket, name)

	r, err := s.client.Bucket(s.bucket).Object(name).NewReader(c)
	if err != nil {
		return nil, err
	}

	data := make([]byte, r.Size())
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
