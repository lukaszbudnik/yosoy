FROM golang:1.17.3-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

# build yosoy
ADD . /go/yosoy
RUN cd /go/yosoy && go build

FROM alpine:3.15.1
COPY --from=builder /go/yosoy/yosoy /bin

# register entrypoint
ENTRYPOINT ["yosoy"]

EXPOSE 80
