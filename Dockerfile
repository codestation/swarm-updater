FROM golang:1.11-alpine as builder

ARG BUILD_NUMBER
ARG BUILD_COMMIT_SHORT
ENV GO111MODULE on
ENV CGO_ENABLED 0

WORKDIR /app
COPY . .

RUN go install -mod vendor -ldflags "-w -s \
   -X main.BuildNumber=${BUILD_NUMBER} \
   -X main.BuildCommit=${BUILD_COMMIT_SHORT} \
  -X \"main.BuildTime=$(date -u '+%Y-%m-%d %I:%M:%S %Z')\"" \
  -a .

FROM alpine:3.9
LABEL maintainer="codestation <codestation404@gmail.com>"
LABEL xyz.megpoid.swarm-updater="true"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /go/bin/swarm-updater /bin/swarm-updater

ENTRYPOINT ["/bin/swarm-updater"]
