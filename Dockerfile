FROM golang:1.19-alpine as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

# build yosoy
ADD . /go/yosoy
RUN cd /go/yosoy && go build

FROM alpine:3.17
COPY --from=builder /go/yosoy/yosoy /bin

# register entrypoint
ENTRYPOINT ["yosoy"]

EXPOSE 80
