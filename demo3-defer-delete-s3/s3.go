package s3

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func createS3Bucket(s3Client *s3.Client, name string, region string) error {
	var lastError error
	for range 3 {
		if _, err := s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
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
			context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(name)}, 5*time.Second); err != nil {
			log.Printf("Failed attempt to wait for bucket %s to exist.\n", name)
			lastError = err
			continue
		}
		return nil
	}
	slog.Error("Failed to create S3 bucket after multiple attempts", "bucket", name, "error", lastError)
	return lastError
}

func deleteBucket(s3Client *s3.Client, name string, region string) error {
	output, err := s3Client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		slog.Error("Failed to delete S3 bucket", "bucket", name, "error", err)
		return err
	}
	slog.Info("S3 bucket deleted successfully", "bucket", name, "output", output.ResultMetadata)
	return nil
}
