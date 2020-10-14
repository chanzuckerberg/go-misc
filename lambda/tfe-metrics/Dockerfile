FROM golang:1.14 AS builder

COPY . /app
WORKDIR /app

# Build the Go app
RUN make build

FROM alpine:latest AS final

RUN apk update && apk --no-cache add ca-certificates curl

COPY --from=builder /app/tfe-metrics /bin/tfe-metrics

ADD https://github.com/segmentio/chamber/releases/download/v2.7.5/chamber-v2.7.5-linux-amd64 /bin/chamber
RUN chmod +x /bin/chamber

CMD ["tfe-metrics"]
