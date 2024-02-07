FROM golang:1.21.6-alpine3.19 AS builder

COPY . .

RUN GOPATH= go build -o /main .

FROM scratch

COPY --from=builder main main

ENTRYPOINT ["/main"]
