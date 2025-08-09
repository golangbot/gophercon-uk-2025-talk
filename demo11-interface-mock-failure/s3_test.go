package s3

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	callCount map[string]int
}

func (m mockS3Client) CreateBucket(ctx context.Context,
	params *s3.CreateBucketInput,
	optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	m.callCount["CreateBucket"] = m.callCount["CreateBucket"] + 1
	if m.callCount["CreateBucket"] <= 2 {
		return nil, errors.New("mocked error: failed to create bucket")
	}
	return &s3.CreateBucketOutput{}, nil
}

func (m mockS3Client) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return &s3.HeadBucketOutput{}, nil
}

func (m mockS3Client) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return &s3.DeleteBucketOutput{}, nil
}

func Test_createS3BucketSuccessfulRetry(t *testing.T) {
	mockS3Client := mockS3Client{
		callCount: make(map[string]int),
	}
	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false
	var testLogs strings.Builder
	w := io.MultiWriter(os.Stdout, &testLogs)
	h := slog.NewTextHandler(w, nil)
	slog.SetDefault(slog.New(h))
	defer deleteBucket(mockS3Client, bucketName, region)
	if err := createS3Bucket(mockS3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}
	if !strings.Contains(testLogs.String(), "Failed to create S3 bucket") {
		t.Errorf("Expected s3 bucket failure but did not find it in logs")
	}
}
