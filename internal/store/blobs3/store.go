package blobs3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type Config struct {
	Region          string
	Bucket          string
	Endpoint        string
	PathStyle       bool
	AccessKeyID     string
	SecretAccessKey string
}

type Store struct {
	client *s3.Client
	bucket string
}

func NewStore(ctx context.Context, cfg Config) (Store, error) {
	if strings.TrimSpace(cfg.Region) == "" {
		return Store{}, errors.New("region is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return Store{}, errors.New("bucket is required")
	}

	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		loadOptions = append(loadOptions, config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AccessKeyID,
				SecretAccessKey: cfg.SecretAccessKey,
			},
		}))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return Store{}, err
	}

	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.UsePathStyle = cfg.PathStyle
		if strings.TrimSpace(cfg.Endpoint) != "" {
			options.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return Store{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s Store) Put(ctx context.Context, objectKey string, data []byte) error {
	if strings.TrimSpace(objectKey) == "" {
		return errors.New("object key is required")
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
	})
	return err
}

func (s Store) Get(ctx context.Context, objectKey string) ([]byte, error) {
	if strings.TrimSpace(objectKey) == "" {
		return nil, errors.New("object key is required")
	}

	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.ErrorCode() {
			case "NoSuchKey", "NotFound":
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
