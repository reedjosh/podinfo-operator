# podinfo-operator
Toy operator that deploys and controls the [podinfo](https://github.com/stefanprodan/podinfo) application.

## Description

Just a practice Kubernetes Operator pattern built using `kubebuilder`.

### An aside about application deployment.

Operators today should not simply deploy a basic application as there are many other alternatives.

For demonstration of a better method, I have self hosted an ArgoCD Application deployment of `podinfo`.

One can login (`readonlyuser`/`readonlypass`) 
[here](https://deploy.thereedfamily.rocks/applications/argocd/podinfo-system?view=tree&resource=)
and see this same application deployed via ArgoCD pointing to [podinfo](https://github.com/stefanprodan/podinfo)'s own
[helm chart](https://github.com/stefanprodan/podinfo/tree/master/charts/podinfo).

The above application is deployed as an example on my home server, and the ArgoCD app in use can be seen
[here](https://git.thereedfamily.rocks/jayr/podinfo-argo/src/branch/main/podinfo.yaml).

### The podinfo-operator

This operator deploys Redis in the same container as the podinfo application when enabled. A future enhancement would
deploy a proper Redis cluster.

This operator also deploys a service in front of the `pondinfo` deployment and it should be possible to port forward
to the service in the same manner as the deployment.

### Manually testing the podinfo-operator

Port forward to the operator.
``` sh
kubectl port-forward deployments/myappresource-sample 9898:9898
```

Verify caching -- the default `myappresource-sample` at [./config/samples/podinfo_v1alpha1_myappresource.yaml]
enables Redis by default.
``` sh
curl -X PUT -d theargument=thevaule  localhost:9898/cache/thekey
curl -X GET -d thearg=thevalue  localhost:9898/cache/thekey
theargument=thevaule%
```

Navigating to `localhost:<forward-port>` should present the podinfo UI. The colors should update via the 
input to the MyAppResource configuation. Consider switching the default to `#b5bd68`.


## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

- (Optionally) [Install Tilt](https://docs.tilt.dev/install.html)

### Tilt

For development, this project can be installed and live updated via Tilt.
For installation only, see the `makefile` documentation below. 

Once the above prerequisites are installed `tilt up` should perform the equivalent of the `makefile` steps below; 
however, it also watches input files for changes and automatically reruns and applies updates.

Tilt also works better with a cluster local docker registry for a speedup. See the `./hack/kind-with-registry.sh`
script for how to create a kind cluster with a local docker registry.

If using a real cluster, edit `allow_k8s_contexts(['kind-kind'<, 'other context'>])` to add your
k8s cluster context to the list Tilt is allowed to operator against.

Tilt also provides a UI for visualization, control, and log viewing. 

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/podinfo:tag
```

**NOTE:** This image ought to be published in the personal registry you specified. 
And it is required to have access to pull the image from the working environment. 
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/podinfo:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin 
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/podinfo:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/podinfo/<tag or branch>/dist/install.yaml
```

## Contributing

Please fork and file a pull request. As the pipeline is further developed, this process will expand.

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024 Joshua Reed.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Future TODOs and Project Findings

### Future tasks

- Make use of owner references. Specifically don't delete if not matching.
- Deploy Redis as its own deployment with a service for comms -- I really wanted to but just ran out of time.
- Test suite build out -- right now its unimpressive.
- Test and Release Pipeline

### Patching instead of Updating

I really wanted to make patching work. It seemed it may be more efficient and was interesting to play with.

``` go
type patchArrayofStringsValue struct {
	Op    string   `json:"op"`
	Path  string   `json:"path"`
	Value []string `json:"value"`
}
type patchStatus struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value podinfov1alpha1.MyAppResourceStatus `json:"value"`
}
type patchBool struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value bool `json:"value"`
}
type pathces []interface{}

func buildPatch(myApp *podinfov1alpha1.MyAppResource) (client.Patch, error) {
	patch := []interface{}{
		patchArrayofStringsValue{
			Op:    "replace",
			Path:  "/metadata/finalizers",
			Value: myApp.GetFinalizers(),
		},
		patchStatus{
			Op:    "add",
			Path:  "/status",
			Value: myApp.Status,
		},
		patchBool{
			Op:    "replace",
			Path:  "/status/ready",
			Value: true,
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return client.RawPatch(types.JSONPatchType, []byte{}), err
	}
	return client.RawPatch(types.JSONPatchType, patchBytes), nil
}
```

But I could not get it to update status.

### controllerutils

Controller utils provides many useful bits. 

I really wanted to use:
``` Go
controllerutils.CreateOrUpdate
```

But when used in place of my more basic logic it just didn't seem to function.
