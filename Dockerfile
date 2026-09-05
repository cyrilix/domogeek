FROM golang:1.27-alpine AS builder

WORKDIR /go/src
ADD . .
RUN CGO_ENABLED=0 go build -mod=vendor -tags netgo -o /go/bin/domogeek ./cmd/domogeek

FROM gcr.io/distroless/static

USER 1234
COPY --from=builder /go/bin/domogeek /go/bin/domogeek
EXPOSE 8080
ENTRYPOINT ["/go/bin/domogeek"]
