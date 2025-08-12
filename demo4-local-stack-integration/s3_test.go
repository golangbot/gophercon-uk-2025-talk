package s3

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func configureToxiProxy(t *testing.T, removeToxicErrCh chan error) {
	toxiClient := toxiproxy.NewClient("localhost:8474")
	_, err := toxiClient.Populate([]toxiproxy.Proxy{{
		Name:     "s3_proxy",
		Listen:   "localhost:8443",
		Upstream: "localhost.localstack.cloud:4566",
		Enabled:  true,
	}})
	if err != nil {
		t.Fatalf("Failed to create toxy proxy: %s", err)
	}

	s3Proxy, err := toxiClient.Proxy("s3_proxy")
	if err != nil {
		t.Fatalf("Failed to get s3_proxy: %s", err)
	}

	_, err = s3Proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
		"latency": 30000,
	})
	if err != nil {
		t.Fatalf("Failed to add toxic: %s", err)
	}

	go func() {
		<-time.After(7 * time.Second)
		err := s3Proxy.RemoveToxic("latency")
		removeToxicErrCh <- err

	}()
}

func Test_createS3BucketSuccessfulRetry(t *testing.T) {
	removeToxicErr := make(chan error)
	configureToxiProxy(t, removeToxicErr)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-west-2"),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					var d net.Dialer
					conn, err := d.DialContext(ctx, network, "localhost:8443")
					if err != nil {
						return nil, err
					}
					return tls.Client(conn, &tls.Config{
						ServerName: "localhost.localstack.cloud",
					}), nil
				},
			},
		}),
	)
	if err != nil {
		slog.Error("Failed to create AWS session", "error", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	bucketName := "gopherconuk-2025-my-new-bucket"
	region := "eu-west-2"
	wantErr := false

	var testLogs strings.Builder
	w := io.MultiWriter(os.Stdout, &testLogs)
	h := slog.NewTextHandler(w, nil)
	slog.SetDefault(slog.New(h))

	//defer deleteBucket(s3Client, bucketName, "eu-west-2")
	if err := createS3Bucket(s3Client, bucketName, region); (err != nil) != wantErr {
		t.Errorf("createS3Bucket() error = %v, wantErr %v", err, wantErr)
	}
	if _, err := s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}); err != nil {
		t.Errorf("Failed to get S3 bucket: %v", err)
	}
	// Wait for the toxic removal goroutine and check for errors
	if err := <-removeToxicErr; err != nil {
		t.Errorf("Failed to remove toxic: %v", err)
	}
	if !strings.Contains(testLogs.String(), "Failed to create S3 bucket") {
		t.Errorf("Expected s3 bucket failure but did not find it in logs")
	}
}
