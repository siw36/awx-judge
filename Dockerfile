FROM golang:alpine AS builder

ARG BUILD_DATE

LABEL maintainer="rhh.klussmann@gmail.com"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="awx-judge"
LABEL org.label-schema.description="AWX Judge"
LABEL org.label-schema.vcs-url="https://github.com/siw36/awx-judge"

WORKDIR /go/src/github.com/siw36/awx-judge

COPY . .

RUN CGO_ENABLED=0 go build ./cmd/awx-judge/


FROM alpine:latest

USER 0

COPY ./web /var/web

RUN chgrp -R 0 /var/web && \
  chmod -R g=u /var/web

# RUN chmod -R 777 /var/web/static/icons

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/siw36/awx-judge/awx-judge /go/bin/awx-judge

USER 9001

ENV AWX_JUDGE_CONFIG_PATH=/var/run/config.yaml

ENTRYPOINT ["/go/bin/awx-judge"]
