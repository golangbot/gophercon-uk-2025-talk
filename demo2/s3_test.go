package main

import (
	"context"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Test_createS3Bucket(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-west-2"),
	)
	if err != nil {
		slog.Error("Failed to create AWS session", "error", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	// bucketPrefix := "gopherconuk-2025"
	// if err := createS3Bucket(*s3Client, bucketPrefix+"my-new-bucket-3", "eu-west-2"); err != nil {
	// 	slog.Error("Failed to create S3 bucket", "error", err)
	// 	return
	// }
	// slog.Info("S3 bucket created successfully")
	type args struct {
		s3Client *s3.Client
		name     string
		region   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create S3 Bucket",
			args: args{
				s3Client: s3Client,
				name:     "gopherconuk-2025-my-new-bucket",
				region:   "eu-west-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createS3Bucket(tt.args.s3Client, tt.args.name, tt.args.region); (err != nil) != tt.wantErr {
				t.Errorf("createS3Bucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
