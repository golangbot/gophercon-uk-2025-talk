package s3

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func createS3Bucket(s3Client *s3.Client, name string, region string) error {
	var lastError error
	retryCount := 3
	for range retryCount {
		func() {
			lastError = nil
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(name),
				CreateBucketConfiguration: &types.CreateBucketConfiguration{
					LocationConstraint: types.BucketLocationConstraint(region),
				},
			}); err != nil {
				slog.Error("Failed to create S3 bucket", "bucket", name, "error", err)
				lastError = err
				return
			}
			if err := s3.NewBucketExistsWaiter(s3Client).Wait(
				ctx, &s3.HeadBucketInput{Bucket: aws.String(name)}, time.Minute); err != nil {
				slog.Error("Failed attempt to wait for bucket to exist.\n", "error", err)
				lastError = err
				return
			}
		}()
		if lastError == nil {
			slog.Info("S3 bucket created successfully", "bucket", name)
			return nil
		}
	}
	slog.Error("Failed to create S3 bucket after multiple attempts", "bucket", name, "error", lastError)
	return lastError
}

func deleteBucket(s3Client *s3.Client, name string, region string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		slog.Error("Failed to delete S3 bucket", "bucket", name, "error", err)
		return err
	}
	slog.Info("S3 bucket deleted successfully", "bucket", name)
	return nil
}
