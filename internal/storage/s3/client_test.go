package s3store

import (
	"context"
	"strings"
	"testing"
)

func TestNewFromEnvAndPresign(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("S3_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("S3_SECRET_ACCESS_KEY", "secret")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("S3_ENDPOINT", "https://example.com")
	t.Setenv("S3_FORCE_PATH_STYLE", "true")
	t.Setenv("S3_BUCKET", "mybucket")

	s3c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv() error = %v", err)
	}

	putURL, headers, err := s3c.PresignPut(context.Background(), "", "test.txt", "text/plain", 60)
	if err != nil {
		t.Fatalf("PresignPut() error = %v", err)
	}
	if putURL == "" || !strings.Contains(putURL, "mybucket") || !strings.Contains(putURL, "test.txt") {
		t.Fatalf("unexpected presigned PUT url: %s", putURL)
	}
	if headers["Content-Type"] != "text/plain" {
		t.Fatalf("unexpected content-type header: %s", headers["Content-Type"])
	}
	if headers["x-amz-acl"] != "private" {
		t.Fatalf("unexpected x-amz-acl header: %s", headers["x-amz-acl"])
	}

	getURL, err := s3c.PresignGet(context.Background(), "", "test.txt", 60)
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	if getURL == "" || !strings.Contains(getURL, "mybucket") || !strings.Contains(getURL, "test.txt") {
		t.Fatalf("unexpected presigned GET url: %s", getURL)
	}
}

func TestNewFromEnvMissingCredentials(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("S3_ACCESS_KEY_ID", "")
	t.Setenv("S3_SECRET_ACCESS_KEY", "")
	if _, err := NewFromEnv(); err == nil {
		t.Fatalf("expected error when credentials are missing")
	}
}

func TestNewFromEnvMissingBucket(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("S3_ACCESS_KEY_ID", "AKIAEXAMPLE")
	t.Setenv("S3_SECRET_ACCESS_KEY", "secret")
	t.Setenv("S3_BUCKET", "")
	if _, err := NewFromEnv(); err == nil {
		t.Fatalf("expected error when bucket is missing")
	}
}
