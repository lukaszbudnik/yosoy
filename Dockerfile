FROM golang:1.13.5-alpine3.10 as builder

MAINTAINER Łukasz Budnik lukasz.budnik@gmail.com

# build yosoy
RUN apk add git
RUN git clone https://github.com/lukaszbudnik/yosoy.git
RUN cd /go/yosoy && go build

FROM alpine:3.10
COPY --from=builder /go/yosoy/yosoy /bin

# register entrypoint
ENTRYPOINT ["yosoy"]

EXPOSE 80
