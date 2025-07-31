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

func Test_createS3BucketSuccessfulRetry(t *testing.T) {
	var testLogs strings.Builder
	w := io.MultiWriter(os.Stdout, &testLogs)
	h := slog.NewTextHandler(w, nil)
	slog.SetDefault(slog.New(h))
	toxiClient := toxiproxy.NewClient("localhost:8474")
	_, err := toxiClient.Populate([]toxiproxy.Proxy{{
		Name:   "aws_proxy",
		Listen: "localhost:26379",

		Upstream: "127.0.0.1:4566",
		Enabled:  true,
	}})
	if err != nil {
		t.Fatalf("Failed to create toxy proxy: %s", err)
	}

	proxy, err := toxiClient.Proxy("aws_proxy")
	if err != nil {
		t.Fatalf("Failed to get aws_proxy: %s", err)
	}

	_, err = proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
		"latency": 7000,
	})
	if err != nil {
		t.Fatalf("Failed to add toxic: %s", err)
	}

	time.Sleep(3 * time.Second)
	removeToxicErr := make(chan error)

	go func() {
		<-time.After(7 * time.Second)
		err := proxy.RemoveToxic("latency")
		removeToxicErr <- err

	}()

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-west-2"),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					var d net.Dialer
					conn, err := d.DialContext(ctx, network, "localhost:26379")
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
			name: "Create S3 Bucket With Retry",
			args: args{
				s3Client: s3Client,
				name:     "gopherconuk-2025-my-new-bucket",
				region:   "eu-west-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer deleteBucket(tt.args.s3Client, tt.args.name, "eu-west-2")
			if err := createS3Bucket(tt.args.s3Client, tt.args.name, tt.args.region); (err != nil) != tt.wantErr {
				t.Errorf("createS3Bucket() error = %v, wantErr %v", err, tt.wantErr)
			}
			if _, err := tt.args.s3Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
				Bucket: aws.String(tt.args.name),
			}); err != nil {
				t.Errorf("Failed to get S3 bucket: %v", err)
			}
			// Wait for the toxic removal goroutine and check for errors
			if err := <-removeToxicErr; err != nil {
				t.Errorf("Failed to remove toxic: %v", err)
			}
		})
		slog.Info("Testcase  logs", "logs", testLogs.String())
		if !strings.Contains(testLogs.String(), "Failed to create S3 bucket") {
			t.Errorf("Expected s3 bucket failure but did not find it in logs")
		}
	}
}
