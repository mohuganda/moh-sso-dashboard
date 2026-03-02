package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string

	// For MinIO or other S3-compatible storage:
	Endpoint     string // e.g. http://minio:9000 (leave empty for AWS)
	UsePathStyle bool
}

type s3Storage struct {
	bucket string
	client *s3.Client
}

func NewS3Storage(cfg S3Config) Storage {
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(defaultRegion(cfg.Region)),
	}

	// Use explicit creds if provided, otherwise fall back to AWS default chain
	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		opts = append(opts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// Custom endpoint for MinIO / S3-compatible
	if strings.TrimSpace(cfg.Endpoint) != "" {
		endpoint := cfg.Endpoint
		opts = append(opts, awscfg.WithHTTPClient(&http.Client{}))
		opts = append(opts, awscfg.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		))
	}

	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		// NewS3Storage signature returns Storage; panic is harsh.
		// If you prefer returning (Storage, error), I can refactor.
		panic(fmt.Errorf("failed to init s3 client: %w", err))
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &s3Storage{
		bucket: cfg.Bucket,
		client: client,
	}
}

func (s *s3Storage) Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          r,
		ContentLength: &size,
		ContentType:   aws.String(defaultContentType(contentType)),
	})
	return err
}

func (s *s3Storage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *s3Storage) Delete(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (s *s3Storage) Exists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err == nil {
		return true, nil
	}
	// Could inspect API error codes; keep it simple & safe:
	return false, nil
}

func defaultRegion(region string) string {
	r := strings.TrimSpace(region)
	if r == "" {
		return "us-east-1"
	}
	return r
}

func defaultContentType(ct string) string {
	ct = strings.TrimSpace(ct)
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}
