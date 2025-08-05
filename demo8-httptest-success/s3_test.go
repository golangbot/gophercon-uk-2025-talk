package s3

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Test_createS3BucketSuccessfulRetry(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-2"),
		config.WithBaseEndpoint(ts.URL),
		config.WithHTTPClient(ts.Client()),
	)
	if err != nil {
		slog.Error("Failed to create AWS session", "error", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)

	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false

	defer deleteBucket(s3Client, bucketName, region)
	if err := createS3Bucket(s3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}
	if _, err := s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}); err != nil {
		t.Errorf("Failed to get S3 bucket: %v", err)
	}

}
