package s3store

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	configs "github.com/Ulpio/vergo/internal/pkg/config"
)

type S3 struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	DefaultBucket string
}

func NewFromConfig(cfgApp configs.Config) (*S3, error) {
	var (
		awsCfg aws.Config
		err    error
	)

	if cfgApp.S3AccessKeyID != "" && cfgApp.S3SecretAccessKey != "" {
		provider := credentials.NewStaticCredentialsProvider(
			cfgApp.S3AccessKeyID,
			cfgApp.S3SecretAccessKey,
			cfgApp.AWSSessionToken, // pode estar vazio
		)
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfgApp.S3Region),
			config.WithCredentialsProvider(provider),
		)
	} else {
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfgApp.S3Region),
		)
	}
	if err != nil {
		return nil, err
	}

	opts := func(o *s3.Options) {
		o.UsePathStyle = cfgApp.S3ForcePathStyle
		if cfgApp.S3Endpoint != "" {
			o.BaseEndpoint = aws.String(cfgApp.S3Endpoint)
		}
	}

	client := s3.NewFromConfig(awsCfg, opts)
	return &S3{
		Client:        client,
		PresignClient: s3.NewPresignClient(client),
		DefaultBucket: cfgApp.S3Bucket,
	}, nil
}

func (s *S3) PresignPut(ctx context.Context, bucket, key, contentType string, expiresSeconds int64) (string, map[string]string, error) {
	if bucket == "" {
		bucket = s.DefaultBucket
	}
	if expiresSeconds <= 0 || expiresSeconds > 3600 {
		expiresSeconds = 300
	}

	in := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         s3types.ObjectCannedACLPrivate,
	}

	ps, err := s.PresignClient.PresignPutObject(ctx, in, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(expiresSeconds) * time.Second
	})
	if err != nil {
		return "", nil, err
	}

	headers := map[string]string{"Content-Type": contentType}
	return ps.URL, headers, nil
}

func (s *S3) PresignGet(ctx context.Context, bucket, key string, expiresSeconds int64) (string, error) {
	if bucket == "" {
		bucket = s.DefaultBucket
	}
	if expiresSeconds <= 0 || expiresSeconds > 3600 {
		expiresSeconds = 300
	}

	in := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	ps, err := s.PresignClient.PresignGetObject(ctx, in, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(expiresSeconds) * time.Second
	})
	if err != nil {
		return "", err
	}
	return ps.URL, nil
}

func (s *S3) DeleteObject(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = s.DefaultBucket
	}
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}
