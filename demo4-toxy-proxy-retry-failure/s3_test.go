package s3

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Test_createS3BucketRetryFailure(t *testing.T) {
	toxiClient := toxiproxy.NewClient("localhost:8474")
	_, err := toxiClient.Populate([]toxiproxy.Proxy{{
		Name:     "s3_proxy",
		Listen:   "localhost:8443",
		Upstream: "s3.eu-west-2.amazonaws.com:443",
		Enabled:  true,
	}})
	if err != nil {
		t.Fatalf("Failed to create Toxiproxy with error %s", err)
	}

	s3Proxy, err := toxiClient.Proxy("s3_proxy")
	if err != nil {
		t.Fatalf("Failed to get s3_proxy: %s", err)
	}
	latencyToxic, err := s3Proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
		"latency": 30000,
	})
	if err != nil {
		t.Fatalf("Failed to add toxic: %s", err)
	}
	t.Logf("Added %s toxic to s3_proxy", latencyToxic.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	region := "eu-west-2"
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					var d net.Dialer
					conn, err := d.DialContext(ctx, network, "localhost:8443")
					if err != nil {
						return nil, err
					}
					return tls.Client(conn, &tls.Config{
						ServerName: "s3.eu-west-2.amazonaws.com",
					}), nil
				},
			},
		}),
	)
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		return
	}
	s3Client := s3.NewFromConfig(cfg)

	bucketName := "gopherconuk-2025-my-new-bucket"
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
