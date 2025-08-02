package s3

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/mock"
)

func Test_createS3BucketSuccess(t *testing.T) {
	mockS3Client := mocks3Client{}
	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false

	defer deleteBucket(&mockS3Client, bucketName, region)
	mockS3Client.On("CreateBucket", mock.Anything, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}).Return(nil, nil)

	mockS3Client.On("DeleteBucket", mock.Anything, mock.Anything).Return(nil, nil)

	mockS3Client.On("HeadBucket", mock.Anything, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}, mock.Anything).Return(&s3.HeadBucketOutput{}, nil)

	if err := createS3Bucket(&mockS3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}
}

func Test_createS3BucketRetrySuccess(t *testing.T) {
	var testLogs strings.Builder
	w := io.MultiWriter(os.Stdout, &testLogs)
	h := slog.NewTextHandler(w, nil)
	slog.SetDefault(slog.New(h))

	mockS3Client := mocks3Client{}
	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false

	defer deleteBucket(&mockS3Client, bucketName, region)
	mockS3Client.On("CreateBucket", mock.Anything, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}).Return(nil, errors.New("mocked error: failed to create bucket")).Twice()

	mockS3Client.On("CreateBucket", mock.Anything, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}).Return(nil, nil).Once()

	mockS3Client.On("DeleteBucket", mock.Anything, mock.Anything).Return(nil, nil)

	mockS3Client.On("HeadBucket", mock.Anything, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}, mock.Anything).Return(&s3.HeadBucketOutput{}, nil)

	if err := createS3Bucket(&mockS3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}

	if !strings.Contains(testLogs.String(), "Failed to create S3 bucket") {
		t.Errorf("Expected s3 bucket failure but did not find it in logs")
	}
}
