package s3

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Client interface {
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	s3.HeadBucketAPIClient
}

func createS3Bucket(s3Client s3Client, name string, region string) error {
	var lastError error
	for range 3 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() //improve this since three contexts are created in a row
		if _, err := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(name),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(region),
			},
		}); err != nil {
			slog.Error("Failed to create S3 bucket", "bucket", name, "error", err)
			lastError = err
			continue
		}
		if err := s3.NewBucketExistsWaiter(s3Client).Wait(
			ctx, &s3.HeadBucketInput{Bucket: aws.String(name)}, 5*time.Second); err != nil {
			slog.Error("Failed to wait for bucket", "bucket", name, "error", err)
			lastError = err
			continue
		}
		return nil
	}
	slog.Error("Failed to create S3 bucket after multiple attempts", "bucket", name, "error", lastError)
	return lastError
}

func deleteBucket(s3Client s3Client, name string, region string) error {
	_, err := s3Client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		slog.Error("Failed to delete S3 bucket", "bucket", name, "error", err)
		return err
	}
	slog.Info("S3 bucket deleted successfully", "bucket", name)
	return nil
}
