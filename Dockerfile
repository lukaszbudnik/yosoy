FROM golang:1.15.7-alpine3.13 as builder

LABEL maintainer="Łukasz Budnik lukasz.budnik@gmail.com"

# build yosoy
RUN apk --update add git
RUN git clone https://github.com/lukaszbudnik/yosoy.git
RUN cd /go/yosoy && go build

FROM alpine:3.13
COPY --from=builder /go/yosoy/yosoy /bin

# register entrypoint
ENTRYPOINT ["yosoy"]

EXPOSE 80
