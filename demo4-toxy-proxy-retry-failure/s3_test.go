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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Test_createS3BucketFailedRetry(t *testing.T) {
	toxiClient := toxiproxy.NewClient("localhost:8474")
	_, err := toxiClient.Populate([]toxiproxy.Proxy{{
		Name:   "aws_proxy",
		Listen: "localhost:26379",

		Upstream: "s3.eu-west-2.amazonaws.com:443",
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
		"latency": 15000,
	})
	if err != nil {
		t.Fatalf("Failed to add toxic: %s", err)
	}

	time.Sleep(3 * time.Second)

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
						InsecureSkipVerify: true,
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
			defer deleteBucket(tt.args.s3Client, tt.args.name, "eu-west-2")
			if err := createS3Bucket(tt.args.s3Client, tt.args.name, tt.args.region); (err != nil) != tt.wantErr {
				t.Errorf("createS3Bucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
