# yosoy ![Go](https://github.com/lukaszbudnik/yosoy/workflows/Go/badge.svg) ![Docker](https://github.com/lukaszbudnik/yosoy/workflows/Docker%20Image%20CI/badge.svg)

yosoy is an HTTP service for stubbing and prototyping distributed applications. It is a service that introduces itself to the caller and prints useful information about its runtime environment. 

yosoy is extremely useful when creating a stub for a distributed application, as it provides more meaningful responses than, for example, a default nginx welcome page. Further, yosoy incorporates a built-in reachability analyzer to facilitate troubleshooting connectivity issues in distributed systems. A dedicated reachability analyzer endpoint validates network connectivity between yosoy and remote endpoints.

Typical use cases include:

- Testing HTTP routing and ingress
- Testing HTTP load balancing
- Testing HTTP caching
- Executing reachability analysis
- Stubbing and prototyping distributed applications

"Yo soy" means "I am" in Spanish.

## API

yosoy responds to all requests with a JSON containing the information about:

- HTTP request:
  - Host
  - Request URI
  - Method
  - Scheme
  - Proto
  - URL
  - Remote IP
  - HTTP headers
  - HTTP proxy headers
- host:
  - Hostname
  - How many times it was called
  - Env variables if `YOSOY_SHOW_ENVS` is set to `true`, `yes`, `on`, or `1`
  - Files' contents if `YOSOY_SHOW_FILES` is set to a comma-separated list of (valid) files

Check [Sample JSON response](#sample-json-response) to see how you can use yosoy for stubbing/prototyping/troubleshooting distributed applications.

## ping/reachability analyzer

yosoy includes a simple ping/reachability analyzer. You can use this functionality when prototyping distributed systems to validate whether a given component can reach a specific endpoint. yosoy exposes a dedicated `/_/yosoy/ping` endpoint which accepts the following 4 query parameters:

* `h` - required - hostname of the endpoint
* `p` - required - port of the endpoint
* `n` - optional - network, all valid Go networks are supported (including the most popular ones like `tcp`, `udp`, IPv4, IPV6, etc.). If `n` parameter is not provided, it defaults to `tcp`. If `n` parameter is set to unknown network, an error will be returned.
* `t` - optional - timeout in seconds. If `t` parameter is not provided, it defaults to `10`. If `t` contains invalid integer literal, an error will be returned.

Check [Sample ping/reachability analyzer responses](#sample-pingreachability-analyzer-responses) to see how you can use yosoy for troubleshooting network connectivity.

## Docker image

The docker image is available on docker hub and ghcr.io:

```sh
docker pull lukasz/yosoy
docker pull ghcr.io/lukaszbudnik/yosoy
```

It exposes HTTP service on port 80.

## Kubernetes example

There is a sample Kubernetes deployment file in the `test` folder. It uses both `YOSOY_SHOW_ENVS` and `YOSOY_SHOW_FILES` features. The deployment uses Kubernetes Downward API to expose labels and annotations as volume files which are then returned by yosoy.

Deploy it to minikube and execute curl to the service a couple of times:

```bash
# start minikube
minikube start
# deploy test service
kubectl apply -f test/deployment.yaml
# tunnel to it and copy the URL as $URL variable
minikube service --url camarero
# simulate some HTTP requests
curl -H "Host: gateway.myapp.com" $URL/camarero/abc
curl -H "Host: gateway.myapp.com" $URL/camarero/abc
curl -H "Host: gateway.myapp.com" $URL/camarero/abc
curl -H "Host: gateway.myapp.com" $URL/camarero/abc
```

## Sample JSON response

A sample yosoy JSON response to a request made from a single page application (SPA) to a backend API deployed in Kubernetes behind nginx ingress and [haproxy-auth-gateway](https://github.com/lukaszbudnik/haproxy-auth-gateway) looks like this:

```json
{
  "host": "api.localtest.me",
  "proto": "HTTP/1.1",
  "method": "GET",
  "scheme": "https",
  "requestUri": "/camarero",
  "url": "https:///camarero",
  "remoteAddr": "192.168.65.3",
  "counter": 1,
  "headers": {
    "Accept": ["*/*"],
    "Accept-Encoding": ["gzip, deflate, br"],
    "Accept-Language": ["en-US,en;q=0.9,pl-PL;q=0.8,pl;q=0.7,es;q=0.6"],
    "Authorization": [
      "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJXejFuaDNCWDI4UHMxVEMzSDRoOW52Q1VWRXpjVVBzQms4Z1NmeEp4ZS1JIn0.eyJleHAiOjE2Mjk4MjM3OTMsImlhdCI6MTYyOTgyMjg5MywiYXV0aF90aW1lIjoxNjI5ODIyODkyLCJqdGkiOiI3ZmQzMjkwZi05NjMyLTQ0NzEtYjRjOS1lNTFjZDYwMjllYjgiLCJpc3MiOiJodHRwczovL2F1dGgubG9jYWx0ZXN0Lm1lL2F1dGgvcmVhbG1zL2hvdGVsIiwic3ViIjoiMDdmYzM3YmYtMmJjNy00ZTRmLWE3MDUtYzRjNjgzNTIwYmU1IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoicmVhY3QiLCJub25jZSI6IjQzNDhmMjU5LTliYTYtNDk2ZC04N2I5LWZmZGYzNDMwN2UzOSIsInNlc3Npb25fc3RhdGUiOiJmNTM5OGI3Ny01OTNhLTQ3OWYtOTc5NS00NGIyNGJjMjhkYjQiLCJhY3IiOiIxIiwiYWxsb3dlZC1vcmlnaW5zIjpbImh0dHBzOi8vbHVrYXN6YnVkbmlrLmdpdGh1Yi5pbyJdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsiY2FtYXJlcm8iXX0sInNjb3BlIjoib3BlbmlkIGVtYWlsIHByb2ZpbGUiLCJzaWQiOiJmNTM5OGI3Ny01OTNhLTQ3OWYtOTc5NS00NGIyNGJjMjhkYjQiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsIm5hbWUiOiJKdWxpbyIsInByZWZlcnJlZF91c2VybmFtZSI6Imp1bGlvIiwiZ2l2ZW5fbmFtZSI6Ikp1bGlvIn0.t5y3L4FzGxM0zwI3fskDI8Kemxz_izcvPPKciSEvNHnZWGQK-9AclGNFz_A9cLFSkpc6l6lBmt7WaC0i04c4h1a9G9AOFImmVXPMPDdTXOQ4aah4CvlN6Fy8ShvSHrQA-wMHEELBpIFsKFx2WP3QHiy27ycr3kqQzW4QymyU7J8tM4-qKR_H1_8aiNOrm5fIED-nEP096V2zvWXiGZX7ts6XE2-annhKphCABLdmIiwgDUnhlAz0hdiDrDbIjzr0ldW4AnUkSQxIZY0PnoEnGVuUvkOYvJpFx10gjORMnRgHSEj9Mk5dtyVGHcihZ5TntCL40WoymNxae6K4-FH3Lw"
    ],
    "Origin": ["https://lukaszbudnik.github.io"],
    "Referer": ["https://lukaszbudnik.github.io/"],
    "Sec-Ch-Ua": [
      "\" Not;A Brand\";v=\"99\", \"Google Chrome\";v=\"91\", \"Chromium\";v=\"91\""
    ],
    "Sec-Ch-Ua-Mobile": ["?0"],
    "Sec-Fetch-Dest": ["empty"],
    "Sec-Fetch-Mode": ["cors"],
    "Sec-Fetch-Site": ["cross-site"],
    "User-Agent": [
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
    ],
    "X-Forwarded-For": ["192.168.65.3", "10.1.3.9"],
    "X-Forwarded-Host": ["api.localtest.me"],
    "X-Forwarded-Port": ["443"],
    "X-Forwarded-Proto": ["https"],
    "X-Real-Ip": ["192.168.65.3"],
    "X-Request-Id": ["48a77564d88ca8a893610b9458bfd885"],
    "X-Scheme": ["https"]
  },
  "hostname": "camarero-cf7c95ccd-cz5lh",
  "envVariables": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "HOSTNAME=camarero-cf7c95ccd-cz5lh",
    "YOSOY_SHOW_FILES=/etc/podinfo/labels,/etc/podinfo/annotations",
    "YOSOY_SHOW_ENVS=true",
    "KUBERNETES_SERVICE_PORT=443",
    "KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443",
    "KUBERNETES_PORT=tcp://10.96.0.1:443",
    "KUBERNETES_PORT_443_TCP_PORT=443",
    "KUBERNETES_SERVICE_HOST=10.96.0.1",
    "KUBERNETES_PORT_443_TCP_PROTO=tcp",
    "KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1",
    "HOME=/root"
  ],
  "files": {
    "/etc/podinfo/annotations": "kubernetes.io/config.seen=\"2021-08-24T15:12:19.555374430Z\"\nkubernetes.io/config.source=\"api\"",
    "/etc/podinfo/labels": "app.kubernetes.io/component=\"api\"\napp.kubernetes.io/name=\"camarero\"\napp.kubernetes.io/part-of=\"hotel\"\napp.kubernetes.io/version=\"0.0.1\"\npod-template-hash=\"cf7c95ccd\""
  }
}
```

## Sample ping/reachability analyzer responses

To test if yosoy can connect to `google.com` on port `443` using default `tcp` network use the following command:

```bash
curl -v "http://localhost/_/yosoy/ping?h=google.com&p=443"
> GET /_/yosoy/ping?h=google.com&p=443 HTTP/1.1
> Host: localhost
> User-Agent: curl/7.86.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Fri, 17 Nov 2023 05:54:36 GMT
< Content-Length: 29
< Content-Type: text/plain; charset=utf-8
<
{"message":"ping succeeded"}
```

To see an unsuccessful response you may use localhost with some random port number:

```bash
curl -v "http://localhost/_/yosoy/ping?h=127.0.0.1&p=12345"
> GET /_/yosoy/ping?h=127.0.0.1&p=12345 HTTP/1.1
> Host: localhost
> User-Agent: curl/7.86.0
> Accept: */*
>
< HTTP/1.1 500 Internal Server Error
< Date: Fri, 17 Nov 2023 05:53:48 GMT
< Content-Length: 66
< Content-Type: text/plain; charset=utf-8
<
{"error":"dial tcp 127.0.0.1:12345: connect: connection refused"}
```

## Building and testing locally

Here are some commands to get you started.

Run yosoy directly on port 80.

```bash
go test -coverprofile cover.out
go tool cover -html=cover.out
go run server.go
```

Building local Docker container and run it on port 8080:

```bash
docker build -t yosoy-local:latest .
docker run --rm --name yosoy-local -p 8080:80 yosoy-local:latest
```