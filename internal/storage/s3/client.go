// internal/storage/s3/client.go
package s3store

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3 struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	DefaultBucket string
}

func NewFromEnv() (*S3, error) {
	region := getenv("S3_REGION", "us-east-2")
	endpoint := os.Getenv("S3_ENDPOINT")
	akid := os.Getenv("S3_ACCESS_KEY_ID")
	secret := os.Getenv("S3_SECRET_ACCESS_KEY")
	session := os.Getenv("AWS_SESSION_TOKEN")
	forcePathStyle := getbool("S3_FORCE_PATH_STYLE", false)
	bucket := os.Getenv("S3_BUCKET")

	var (
		cfg aws.Config
		err error
	)

	if akid != "" && secret != "" {
		provider := credentials.NewStaticCredentialsProvider(akid, secret, session)
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(provider),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
		)
	}
	if err != nil {
		return nil, err
	}

	opts := func(o *s3.Options) {
		o.UsePathStyle = forcePathStyle
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}

	client := s3.NewFromConfig(cfg, opts)
	return &S3{
		Client:        client,
		PresignClient: s3.NewPresignClient(client),
		DefaultBucket: bucket,
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

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func getbool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
