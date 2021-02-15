#!/usr/bin/env bash
docker rm awx-judge
docker run --name awx-judge \
  -p 8080:8080/tcp \
  --net my-mongo-cluster \
  --mount type=bind,source="$(pwd)"/configs/config.yaml,target=/var/run/config.yaml \
  awx-judge:latest
