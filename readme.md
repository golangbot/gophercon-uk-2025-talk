### Toxi Proxy
docker run -d --network=host --rm -it shopify/toxiproxy:2.1.4

curl -v http://localhost:8474/proxies

docker run \
  --rm -it \
  -p 127.0.0.1:4566:4566 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  localstack/localstack:4.7.0


aws --endpoint-url=http://localhost:4566 s3 ls

localstack is not a self signed cert, they own the domain and point the dns to 127.0.0.1

so aws --endpoint-url=https://localhost.localstack.cloud:4566 s3 ls works too


go install github.com/vektra/mockery/v3@v3.5.1

mockery