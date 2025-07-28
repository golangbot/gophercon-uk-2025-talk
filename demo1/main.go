package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func createS3Bucket(s3Client s3.Client, name string, region string) error {
	if _, err := s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}); err != nil {
		return err
	}
	if err := s3.NewBucketExistsWaiter(&s3Client).Wait(
		context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(name)}, time.Minute); err != nil {
		log.Printf("Failed attempt to wait for bucket %s to exist.\n", name)
		return err
	}
	return nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-west-2"),
	)
	if err != nil {
		slog.Error("Failed to create AWS session", "error", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	bucketPrefix := "gopherconuk-2025"
	for range 3 {
		if err := createS3Bucket(*s3Client, bucketPrefix+"my-new-bucket-3", "eu-west-2"); err != nil {
			slog.Error("Failed to create S3 bucket", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		slog.Info("S3 bucket created successfully")
		break
	}
}
