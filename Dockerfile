FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/klaudia-sync ./cmd/klaudia-sync

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /workspace

COPY --from=build /out/klaudia-sync /usr/local/bin/klaudia-sync

ENTRYPOINT ["klaudia-sync"]
