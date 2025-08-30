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
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types" // 👈 importa types
)

type S3 struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	DefaultBucket string
}

func NewFromEnv() (*S3, error) {
	region := getenv("S3_REGION", "us-east-1")
	endpoint := os.Getenv("S3_ENDPOINT")
	akid := os.Getenv("S3_ACCESS_KEY_ID")
	secret := os.Getenv("S3_SECRET_ACCESS_KEY")
	forcePathStyle := getbool("S3_FORCE_PATH_STYLE", false)
	bucket := os.Getenv("S3_BUCKET")

	var cfg aws.Config
	var err error
	if akid != "" && secret != "" {
		cfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(akid, secret, "")),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
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
		PresignClient: s3.NewPresignClient(client), // ✅ presigner correto
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

	req := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         s3types.ObjectCannedACLPrivate, // ✅ agora compila
	}

	ps, err := s.PresignClient.PresignPutObject(ctx, req, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(expiresSeconds) * time.Second
	})
	if err != nil {
		return "", nil, err
	}

	// o cliente deve enviar o mesmo Content-Type no PUT
	headers := map[string]string{
		"Content-Type": contentType,
	}
	return ps.URL, headers, nil
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
