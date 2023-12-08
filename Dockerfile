FROM golang:1.21-alpine as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

# install prerequisites
RUN apk update && apk add git

# build yosoy
ADD . /go/yosoy
RUN go env -w GOPROXY=direct
RUN cd /go/yosoy && go build

FROM alpine:3.19
COPY --from=builder /go/yosoy/yosoy /bin

# register entrypoint
ENTRYPOINT ["yosoy"]

EXPOSE 80
