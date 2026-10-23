package s3_test

import (
	"testing"

	"code-base-golang/internal/adapters/s3"
	"code-base-golang/internal/config"
)

func TestS3_New_ValidationErrors(t *testing.T) {
	// Empty endpoint
	_, err := s3.New(config.Config{
		S3Endpoint:  "",
		S3AccessKey: "key",
		S3SecretKey: "secret",
	})
	if err == nil {
		t.Error("expected error for empty endpoint, got nil")
	}

	// Empty access key
	_, err = s3.New(config.Config{
		S3Endpoint:  "localhost:9000",
		S3AccessKey: "",
		S3SecretKey: "secret",
	})
	if err == nil {
		t.Error("expected error for empty access key, got nil")
	}

	// Empty secret key
	_, err = s3.New(config.Config{
		S3Endpoint:  "localhost:9000",
		S3AccessKey: "key",
		S3SecretKey: "",
	})
	if err == nil {
		t.Error("expected error for empty secret key, got nil")
	}
}

func TestS3_New_UnreachableEndpoint(t *testing.T) {
	cfg := config.Config{
		S3Endpoint:   "127.0.0.1:59998",
		S3AccessKey:  "minioadmin",
		S3SecretKey:  "minioadmin",
		S3BucketName: "test-bucket",
	}

	client, err := s3.New(cfg)
	if err == nil {
		t.Fatal("expected error connecting to unreachable S3 endpoint, got nil")
	}
	if client != nil {
		t.Fatal("expected client to be nil on failed connection")
	}
}
