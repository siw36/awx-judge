FROM golang:1.15 AS builder
LABEL maintainer="rhh.klussmann@gmail.com"

ARG BUILD_DATE

LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="awx-judge"
LABEL org.label-schema.description="AWX Judge"
LABEL org.label-schema.vcs-url="https://github.com/siw36/awx-judge"

WORKDIR /go/src/awx-judge

COPY . .

RUN go get -v \
  github.com/Sirupsen/logrus \
  github.com/hashicorp/go-retryablehttp \
  github.com/google/uuid \
  go.mongodb.org/mongo-driver/mongo \
  github.com/kelseyhightower/envconfig \
  gopkg.in/yaml.v2 \
  github.com/gorilla/mux \
  github.com/gorilla/securecookie

RUN go build -o ./awx-judge



FROM scratch

ARG BUILD_DATE
LABEL maintainer="rhh.klussmann@gmail.com"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="awx-judge"
LABEL org.label-schema.description="AWX Judge"
LABEL org.label-schema.vcs-url="https://github.com/siw36/awx-judge"

USER 9001

COPY --from=builder /go/src/app/awx-judge /go/bin/awx-judge

ENTRYPOINT ["/go/bin/awx-judge"]
