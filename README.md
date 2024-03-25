# podinfo
Toy operator that deploys and controls the [podinfo](https://github.com/stefanprodan/podinfo) application.

## Description

Just a practice Kubernetes Operator pattern built using `kubebuilder`.

Operators today should not simply deploy a basic application as there are many other alternatives.

For demonstration of a better method, I have self hosted an ArgoCD Application deployment of `podinfo`.

One can login (`readonlyuser`/`readonlypass`) 
[here](https://deploy.thereedfamily.rocks/applications/argocd/podinfo-system?view=tree&resource=)
and see this same application deployed via ArgoCD pointing to [podinfo](https://github.com/stefanprodan/podinfo)'s own
[helm chart](https://github.com/stefanprodan/podinfo/tree/master/charts/podinfo).

The above application is deployed as an example on my home server, and the ArgoCD app in use can be seen
[here](https://git.thereedfamily.rocks/jayr/podinfo-argo/src/branch/main/podinfo.yaml).

## Getting Started

### Prerequisites
- go version v1.21.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

- (Optionally) [Install Tilt](https://docs.tilt.dev/install.html)

### Tilt

For installation only, see the `makefile` documentation below. 
For development, this project can be installed and live updated via Tilt.

Once the above prerequisites are installed `tilt up` should perform the equivalent of the `makefile` steps below; 
however, it also watches input files for changes and automatically reruns and applies on said changes.

It also provides a nice UI for visualization, control, and 

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

