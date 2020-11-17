# yosoy

yosoy is a HTTP service for stubbing and prototyping distributed applications. It is a service which will introduce itself to the caller and print some useful information about its environment. "Yo soy" in español means "I am".

yosoy is extremely useful when creating a distributed application stub and you need to see a more meaningful responses than a default nginx welcome page.

Typical use cases include:

* testing HTTP routing & ingress
* testing HTTP load balancing
* testing HTTP caching
* stubbing and prototyping the infrastructure

yosoy will provide information like:

* Request URI
* Hostname
* Remote IP
* How many times it was called
* HTTP headers
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

Let's take a look at a sample Kubernetes deployment file. It uses both `YOSOY_SHOW_ENVS` and `YOSOY_SHOW_FILES`.

> To illustrate `YOSOY_SHOW_FILES` functionality it uses Kubernetes Downward API to expose labels and annotations as volume files which are then read by yosoy.

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: camarero
  labels:
    app.kubernetes.io/name: camarero
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: camarero
  template:
    metadata:
      labels:
        app.kubernetes.io/name: camarero
    spec:
      containers:
      - name: yosoy
        image: lukasz/yosoy
        env:
          - name: YOSOY_SHOW_ENVS
            value: "true"
          - name: YOSOY_SHOW_FILES
            value: "/etc/podinfo/labels,/etc/podinfo/annotations"
        ports:
        - containerPort: 80
        volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
      volumes:
        - name: podinfo
          downwardAPI:
            items:
              - path: "labels"
                fieldRef:
                  fieldPath: metadata.labels
              - path: "annotations"
                fieldRef:
                  fieldPath: metadata.annotations
---
apiVersion: v1
kind: Service
metadata:
  name: camarero
  labels:
    app.kubernetes.io/name: camarero
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: camarero
  ports:
    - protocol: TCP
      port: 80
```

Deploy above service (with 2 replicas) and execute curl to the service a couple of times:

```
kubectl apply -f test-deployment.yaml
export NODE_PORT=$(kubectl get services/camarero -o go-template='{{(index .spec.ports 0).nodePort}}')
curl $(minikube ip):$NODE_PORT
curl $(minikube ip):$NODE_PORT
curl $(minikube ip):$NODE_PORT
curl $(minikube ip):$NODE_PORT
```

A sample response looks like this:

```
Request URI: /
Hostname: camarero-859d7c6d6b-kb5s5
Remote IP: 172.18.0.1
Called: 2

HTTP headers:
User-Agent: curl/7.58.0
Accept: */*

Env variables:
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOSTNAME=camarero-859d7c6d6b-kb5s5
YOSOY_SHOW_ENVS=true
YOSOY_SHOW_FILES=/etc/podinfo/labels,/etc/podinfo/annotations
CAMARERO_PORT_80_TCP_PORT=80
CAMARERO_PORT_80_TCP_ADDR=10.105.203.131
KUBERNETES_PORT=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP_PORT=443
CAMARERO_SERVICE_HOST=10.105.203.131
KUBERNETES_PORT_443_TCP_PROTO=tcp
KUBERNETES_SERVICE_HOST=10.96.0.1
KUBERNETES_SERVICE_PORT=443
KUBERNETES_SERVICE_PORT_HTTPS=443
KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1
CAMARERO_PORT=tcp://10.105.203.131:80
CAMARERO_SERVICE_PORT=80
CAMARERO_PORT_80_TCP=tcp://10.105.203.131:80
CAMARERO_PORT_80_TCP_PROTO=tcp
HOME=/root

File /etc/podinfo/labels:
app.kubernetes.io/name="camarero"
pod-template-hash="859d7c6d6b"

File /etc/podinfo/annotations:
kubernetes.io/config.seen="2020-11-17T07:38:15.374049163Z"
kubernetes.io/config.source="api"
```
