package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *S3Storage) Upload(
	ctx context.Context,
	objectKey string,
	r io.Reader,
	size int64,
	contentType string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	return err
}

func (s *S3Storage) Download(
	ctx context.Context,
	objectKey string,
) (io.ReadCloser, error) {

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}

	return out.Body, nil
}

func (s *S3Storage) Delete(
	ctx context.Context,
	objectKey string,
) error {

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (s *S3Storage) Exists(
	ctx context.Context,
	objectKey string,
) (bool, error) {

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return false, nil
	}

	return true, nil
}
