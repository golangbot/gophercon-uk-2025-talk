### Climbing the Testing Pyramid: From Real Service to Interface Mocks in Go
This repository contains the source code and presentation used for my GopherCon UK 2025 talk "Climbing the Testing Pyramid: From Real Service to Interface Mocks in Go"

### Run Toxi Proxy
docker run -d --network=host --rm -it shopify/toxiproxy:2.1.4

#### List proxies in Toxi Proxy
curl -v http://localhost:8474/proxies

### Run local stack
docker run -d \
  --rm -it \
  -p 127.0.0.1:4566:4566 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  localstack/localstack:4.7.0

#### List containers in Local Stack
aws --endpoint-url=https://localhost.localstack.cloud:4566 s3 ls

### Install mockery
go install github.com/vektra/mockery/v3@v3.5.1

#### Run mockery to generate mocks
mockery