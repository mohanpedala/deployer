# deployer
Controller that works with the custom resource which deploys an application and its data store inside kubernetes

## Required Tools
* Go
* Docker desktop
* Enable Kubernetes in Docker-Desktop



### Prerequisites
- go version v1.21.0+
- Docker Desktop 4.27.0
- K8s Client Version: v1.29.1
- K8s Server Version: v1.29.1

### Procedure to run the CRD
**Clone and Navigate to the repo**
* Clone repository
```sh
git clone <https://>/<git>@github.com:mohanpedala/deployer.git
```
* Navigate to repository locally
```
cd ~/deployer
```
**Generate Manifests:**

```sh
make manifests
```
**Install the CRDs into the cluster:**

```sh
make install
```
**Run CRD:**
```sh
make run
```
**Deploy the Custom Resource:**

***Note:*** Enable Redis to test the complete scenario
```sh
cat <<EOF | kubectl apply -f -
apiVersion: mohanbinarybutter.mohanbinarybutter.com/v1alpha1
kind: MyAppResource
metadata:
  name: hello
spec:
  replicaCount: 2
  resources:
    memoryLimit: 64Mi
    cpuRequest: 100m
  image:
    repository: ghcr.io/stefanprodan/podinfo
    tag: latest
  ui:
    color: "#34577c"
    message: "some string"
  redis:
    enabled: true
EOF
```

**Port Forward:**
```sh
kubectl port-forward deploy/hello-deployment 8080:9898
```

**Application Verification:**
```sh
## Post
curl -X POST http://localhost:8080/cache/samplekey -d '{"sample": "samplekeybyme"}' -H "Content-Type: application/json"

## Verify the key
curl http://localhost:8080/cache/samplekey

## Update
curl -X PUT http://localhost:8080/cache/samplekey -d '{"sample": "newsamplekeybyme"}' -H "Content-Type: application/json"

## Verify the new key
curl http://localhost:8080/cache/samplekey
```

**Clean Up:**
* Delete Custom resource manifest
```sh
kubectl delete -f myAppResource.yaml
```
* Break the running code
```sh
ctrl + d
```
* Delete the CRD
```sh
kubectl delete crd myappresources.mohanbinarybutter.mohanbinarybutter.com
```