package s3

import (
	"testing"
)

func Test_createS3BucketSuccessfulRetry(t *testing.T) {
	mockS3Client := mockS3Client{}
	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false

	defer deleteBucket(mockS3Client, bucketName, region)
	if err := createS3Bucket(mockS3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}
}
