FROM golang:1.16-alpine as builder

ARG CI_TAG
ARG BUILD_NUMBER
ARG BUILD_COMMIT_SHORT
ARG CI_BUILD_CREATED
ENV GO111MODULE on

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o release/swarm-updater \
   -ldflags "-w -s \
   -X main.Version=${CI_TAG} \
   -X main.BuildNumber=${BUILD_NUMBER} \
   -X main.Commit=${BUILD_COMMIT_SHORT} \
   -X main.BuildTime=${CI_BUILD_CREATED}"

FROM alpine:3.13
LABEL maintainer="codestation <codestation404@gmail.com>"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/release/swarm-updater /bin/swarm-updater

ENTRYPOINT ["/bin/swarm-updater"]
