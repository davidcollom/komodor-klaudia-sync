FROM alpine:3.21

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates

WORKDIR /workspace

# GoReleaser provides the prebuilt binary artifact in the docker build context.
COPY ${TARGETPLATFORM}/klaudia-sync /usr/local/bin/klaudia-sync

ENTRYPOINT ["klaudia-sync"]
