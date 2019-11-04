FROM golang:alpine as builder

WORKDIR /go/src
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -tags netgo -o /go/bin/domogeek cmd/domogeek/domogeek.go




FROM gcr.io/distroless/static

USER 1234
COPY --from=builder /go/bin/domogeek /go/bin/domogeek
EXPOSE 8080
ENTRYPOINT ["/go/bin/domogeek"]
