FROM golang:1.10 as builder
WORKDIR /go/src/github.com/stelligent/mu/
ADD . .
RUN make deps
RUN make test
RUN make build

FROM docker:stable
COPY --from=builder /go/src/github.com/stelligent/mu/dist/linux_amd64/mu /usr/bin/mu
ENTRYPOINT ["/usr/bin/mu"]