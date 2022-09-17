# enable-long-processing-k8s
MSc project with RedHat
## Installation
This project can fully run locally and includes automation to deploy a local Kubernetes cluster (using Kind).

### Requirements
* Docker
* kubectl
* [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* Go >=1.16 (optional)

## Usage
### Create Cluster
First, we need to create a Kubernetes cluster:
```
❯ make cluster

🔧 Creating Kubernetes cluster...
kind create cluster --config dev/manifests/kind/kind.cluster.yaml
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.25.0-beta.0) 🖼
 ✓ Preparing nodes 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Thanks for using kind! 😊
```

Make sure that the Kubernetes node is ready:
```
❯ kubectl get nodes
NAME                 STATUS   ROLES                  AGE     VERSION
kind-control-plane   Ready    control-plane,master   3m25s   v1.21.1
```

And that system pods are running happily:
```
❯ kubectl -n kube-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
coredns-558bd4d5db-thwvj                     1/1     Running   0          3m39s
coredns-558bd4d5db-w85ks                     1/1     Running   0          3m39s
etcd-kind-control-plane                      1/1     Running   0          3m56s
kindnet-84slq                                1/1     Running   0          3m40s
kube-apiserver-kind-control-plane            1/1     Running   0          3m54s
kube-controller-manager-kind-control-plane   1/1     Running   0          3m56s
kube-proxy-4h6sj                             1/1     Running   0          3m40s
kube-scheduler-kind-control-plane            1/1     Running   0          3m54s
```

### Deploy Admission Webhook
To configure the cluster to use the admission webhook and to deploy said webhook, simply run:
```
❯ make deploy

📦 Building simple-kubernetes-webhook Docker image...
DOCKER_BUILDKIT=1 docker build -t simple-kubernetes-webhook:latest .
[+] Building 10.2s (12/12) FINISHED                                                                                                                                                                                            
 => [internal] load build definition from Dockerfile                                                                                                                                                                      0.0s
 => => transferring dockerfile: 38B                                                                                                                                                                                       0.0s
 => [internal] load .dockerignore                                                                                                                                                                                         0.0s
 => => transferring context: 2B                                                                                                                                                                                           0.0s
 => resolve image config for docker.io/docker/dockerfile:experimental                                                                                                                                                     1.9s
 => CACHED docker-image://docker.io/docker/dockerfile:experimental@sha256:600e5c62eedff338b3f7a0850beb7c05866e0ef27b2d2e8c02aa468e78496ff5                                                                                0.0s
 => [internal] load metadata for docker.io/library/golang:1.19                                                                                                                                                            1.0s
 => [build 1/4] FROM docker.io/library/golang:1.19@sha256:2d17ffd12a2cdb25d4a633ad25f8dc29608ed84f31b3b983427d825280427095                                                                                                0.0s
 => [internal] load build context                                                                                                                                                                                         0.0s
 => => transferring context: 23.14kB                                                                                                                                                                                      0.0s
 => CACHED [build 2/4] WORKDIR /work                                                                                                                                                                                      0.0s
 => [build 3/4] COPY . /work                                                                                                                                                                                              0.0s
 => [build 4/4] RUN --mount=type=cache,target=/root/.cache/go-build,sharing=private   go build -o bin/admission-webhook .                                                                                                 6.3s
 => [run 1/1] COPY --from=build /work/bin/admission-webhook /usr/local/bin/                                                                                                                                               0.1s
 => exporting to image                                                                                                                                                                                                    0.2s
 => => exporting layers                                                                                                                                                                                                   0.2s
 => => writing image sha256:af590c25f6927bec17c554d81ccd7634318270bcbbf492657e98d9463a54cc17                                                                                                                              0.0s
 => => naming to docker.io/library/simple-kubernetes-webhook:latest                                                                                                                                                       0.0s

📦 Pushing admission-webhook image into Kind's Docker daemon...
kind load docker-image simple-kubernetes-webhook:latest
Image: "simple-kubernetes-webhook:latest" with ID "sha256:af590c25f6927bec17c554d81ccd7634318270bcbbf492657e98d9463a54cc17" not yet present on node "kind-control-plane", loading...

♻️  Deleting simple-kubernetes-webhook deployment if existing...
kubectl delete -f dev/manifests/webhook/ || true
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.deploy.yaml": deployments.apps "simple-kubernetes-webhook" not found
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.svc.yaml": clusterroles.rbac.authorization.k8s.io "simple-kubernetes-webhook-getter-cr" not found
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.svc.yaml": serviceaccounts "simple-kubernetes-webhook-sa" not found
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.svc.yaml": clusterrolebindings.rbac.authorization.k8s.io "simple-kubernetes-webhook-getter-crb" not found
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.svc.yaml": services "simple-kubernetes-webhook" not found
Error from server (NotFound): error when deleting "dev/manifests/webhook/webhook.tls.secret.yaml": secrets "simple-kubernetes-webhook-tls" not found

⚙️  Applying cluster config...
kubectl apply -f dev/manifests/cluster-config/
namespace/apps created
mutatingwebhookconfiguration.admissionregistration.k8s.io/simple-kubernetes-webhook.acme.com created
validatingwebhookconfiguration.admissionregistration.k8s.io/simple-kubernetes-webhook.acme.com created

🚀 Deploying simple-kubernetes-webhook...
kubectl apply -f dev/manifests/webhook/
deployment.apps/simple-kubernetes-webhook created
clusterrole.rbac.authorization.k8s.io/simple-kubernetes-webhook-getter-cr created
serviceaccount/simple-kubernetes-webhook-sa created
clusterrolebinding.rbac.authorization.k8s.io/simple-kubernetes-webhook-getter-crb created
service/simple-kubernetes-webhook created
secret/simple-kubernetes-webhook-tls created

```

Then, make sure the admission webhook pod is running (in the `default` namespace):
```
❯ kubectl get pods
NAME                                        READY   STATUS    RESTARTS   AGE
simple-kubernetes-webhook-77444566b7-wzwmx   1/1     Running   0          2m21s
```

You can stream logs from it:
```
❯ make logs

🔍 Streaming simple-kubernetes-webhook logs...
kubectl logs -l app=simple-kubernetes-webhook -f
time="2021-09-03T04:59:10Z" level=info msg="Listening on port 443..."
time="2021-09-03T05:02:21Z" level=debug msg=healthy uri=/health
```

And hit it's health endpoint from your local machine:
```
❯ curl -k https://localhost:8443/health
OK
```

### Deploying pods
Deploy a valid test pod that gets successfully created:
```
❯  make pod

🚀 Deploying test pod...
kubectl apply -f dev/manifests/pods/lifespan-seven.pod.yaml
pod/lifespan-seven created

```

### Deleting pods

```
make delete-pod

♻️ Deleting test pod...
kubectl delete -f dev/manifests/pods/lifespan-seven.pod.yaml
pod "lifespan-seven" deleted

```

