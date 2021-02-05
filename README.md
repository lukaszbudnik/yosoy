# yosoy ![Go](https://github.com/lukaszbudnik/yosoy/workflows/Go/badge.svg) ![Docker](https://github.com/lukaszbudnik/yosoy/workflows/Docker%20Image%20CI/badge.svg)

yosoy is a HTTP service for stubbing and prototyping distributed applications. It is a service which will introduce itself to the caller and print some useful information about its environment. "Yo soy" in español means "I am".

yosoy is extremely useful when creating a distributed application stub and you need to see a more meaningful responses than a default nginx welcome page.

Typical use cases include:

* testing HTTP routing & ingress
* testing HTTP load balancing
* testing HTTP caching
* stubbing and prototyping distributed applications

## API

yosoy responds to all requests with a JSON containing the information about:

* HTTP request:
  * Host
  * Request URI
  * Remote IP
  * HTTP headers
  * HTTP proxy headers
* host:
  * Hostname
  * How many times it was called
  * Env variables if `YOSOY_SHOW_ENVS` is set to `true`, `yes`, `on`, or `1`
  * Files' contents if `YOSOY_SHOW_FILES` is set to a comma-separated list of (valid) files

See [Kubernetes example](#kubernetes-example) below.

## Docker image

The docker image is available on docker hub:

```
lukasz/yosoy
```

It exposes HTTP service on port 80.

## Kubernetes example

There is a sample Kubernetes deployment file in the `test` folder. It uses both `YOSOY_SHOW_ENVS` and `YOSOY_SHOW_FILES`. The deployment uses Kubernetes Downward API to expose labels and annotations as volume files which are then returned by yosoy.

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

A sample response looks like this:

```json
{
  "host": "gateway.myapp.com",
  "requestUri": "/camarero/abc",
  "remoteAddr": "172.17.0.1",
  "counter": 4,
  "headers": {
    "Accept": [
      "*/*"
    ],
    "User-Agent": [
      "curl/7.64.1"
    ]
  },
  "hostname": "camarero-77787464ff-hjdkq",
  "envVariables": [
    "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "HOSTNAME=camarero-77787464ff-hjdkq",
    "YOSOY_SHOW_ENVS=true",
    "YOSOY_SHOW_FILES=/etc/podinfo/labels,/etc/podinfo/annotations",
    "CAMARERO_SERVICE_HOST=10.97.113.33",
    "CAMARERO_PORT=tcp://10.97.113.33:80",
    "CAMARERO_PORT_80_TCP=tcp://10.97.113.33:80",
    "CAMARERO_PORT_80_TCP_ADDR=10.97.113.33",
    "CAMARERO_SERVICE_PORT=80",
    "CAMARERO_PORT_80_TCP_PROTO=tcp",
    "CAMARERO_PORT_80_TCP_PORT=80",
    "HOME=/root"
  ],
  "files": {
    "/etc/podinfo/annotations": "kubernetes.io/config.seen=\"2021-02-03T13:18:34.563751030Z\"\nkubernetes.io/config.source=\"api\"",
    "/etc/podinfo/labels": "app.kubernetes.io/name=\"camarero\"\npod-template-hash=\"77787464ff\""
  }
}
```
