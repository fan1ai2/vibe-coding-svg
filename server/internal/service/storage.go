package service

import (
	"context"
	"io"
	"time"

	"github.com/fan1ai2/vibe-coding-svg/server/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{client: client}, nil
}

func (s *Storage) EnsureBucket(name string) error {
	exists, err := s.client.BucketExists(context.Background(), name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(context.Background(), name, minio.MakeBucketOptions{})
}

func (s *Storage) Upload(bucket, key, contentType string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(context.Background(), bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Download(bucket, key string) (io.ReadCloser, error) {
	return s.client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) PresignedGetURL(bucket, key string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(context.Background(), bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
