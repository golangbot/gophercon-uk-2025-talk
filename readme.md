docker run --network=host --rm -it shopify/toxiproxy:2.1.4


docker run \
  --rm -it \
  -p 127.0.0.1:4566:4566 \
  -p 127.0.0.1:4510-4559:4510-4559 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  localstack/localstack:4.7.0


aws --endpoint-url=http://localhost:4566 s3 ls