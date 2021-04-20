FROM golang:1.16-alpine as builder

ARG CI_COMMIT_TAG
ARG SOURCE_BRANCH
ARG SOURCE_COMMIT
ARG CI_PIPELINE_CREATED_AT
ARG GOPROXY
ENV GOPROXY=${GOPROXY}

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o release/swarm-updater \
   -ldflags "-w -s \
   -X main.Version=${CI_COMMIT_TAG:-$SOURCE_BRANCH} \
   -X main.Commit=${SOURCE_COMMIT:0:8} \
   -X main.BuildTime=${CI_PIPELINE_CREATED_AT:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

FROM alpine:3.13
LABEL maintainer="codestation <codestation404@gmail.com>"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/release/swarm-updater /bin/swarm-updater

ENTRYPOINT ["/bin/swarm-updater"]
