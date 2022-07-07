FROM golang:1.18-alpine as builder

ARG CI_COMMIT_TAG
ARG CI_COMMIT_BRANCH
ARG CI_COMMIT_SHA
ARG CI_PIPELINE_CREATED_AT
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o release/swarm-updater \
   -ldflags "-w -s \
   -X main.Version=${CI_COMMIT_TAG:-$CI_COMMIT_BRANCH} \
   -X main.Commit=${CI_COMMIT_SHA:0:8} \
   -X main.BuildTime=${CI_PIPELINE_CREATED_AT}"

FROM alpine:3.15
LABEL maintainer="codestation <codestation404@gmail.com>"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/release/swarm-updater /bin/swarm-updater

ENTRYPOINT ["/bin/swarm-updater"]
