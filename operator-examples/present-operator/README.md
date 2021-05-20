# Present operator

This is a sample operator using Kubebuilder tool, so generating code, creating CRD YAMLs and deploying code is done in the standard way.

Refer to [present operator](https://github.com/martonsereg/present-operator) 

## Create a kubebuilder porject

```
mkdir present-operator
cd present-operator
go mod init presentation
kubebuilder init --domain meetup.com
```

Then your project struct tree looks like:

```
.
├── bin
│    └── manager
├── config
│    ├── certmanager
│    │    ├── certificate.yaml
│    │    ├── kustomization.yaml
│    │    └── kustomizeconfig.yaml
│    ├── default
│    │    ├── kustomization.yaml
│    │    ├── manager_auth_proxy_patch.yaml
│    │    ├── manager_webhook_patch.yaml
│    │    └── webhookcainjection_patch.yaml
│    ├── manager
│    │    ├── kustomization.yaml
│    │    └── manager.yaml
│    ├── prometheus
│    │    ├── kustomization.yaml
│    │    └── monitor.yaml
│    ├── rbac
│    │    ├── auth_proxy_client_clusterrole.yaml
│    │    ├── auth_proxy_role_binding.yaml
│    │    ├── auth_proxy_role.yaml
│    │    ├── auth_proxy_service.yaml
│    │    ├── kustomization.yaml
│    │    ├── leader_election_role_binding.yaml
│    │    ├── leader_election_role.yaml
│    │    └── role_binding.yaml
│    └── webhook
│        ├── kustomization.yaml
│        ├── kustomizeconfig.yaml
│        └── service.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── hack
│    └── boilerplate.go.txt
├── main.go
├── Makefile
└── PROJECT
```

## Add an API

```
kubebuilder create api --group example --version v1alpha1 --kind Presentation
```
Then run:
```
make manifests
```
This will generate manifests e.g. CRD, RBAC etc.

Now your project struct tree looks like:

```
.
├── api
│    └── v1alpha1
│        ├── groupversion_info.go
│        ├── presentation_types.go
│        └── zz_generated.deepcopy.go
├── bin
│    └── manager
├── config
│    ├── certmanager
│    │    ├── certificate.yaml
│    │    ├── kustomization.yaml
│    │    └── kustomizeconfig.yaml
│    ├── crd
│    │    ├── bases
│    │    │    └── example.meetup.com_presentations.yaml
│    │    ├── kustomization.yaml
│    │    ├── kustomizeconfig.yaml
│    │    └── patches
│    │        ├── cainjection_in_presentations.yaml
│    │        └── webhook_in_presentations.yaml
│    ├── default
│    │    ├── kustomization.yaml
│    │    ├── manager_auth_proxy_patch.yaml
│    │    ├── manager_webhook_patch.yaml
│    │    └── webhookcainjection_patch.yaml
│    ├── manager
│    │    ├── kustomization.yaml
│    │    └── manager.yaml
│    ├── prometheus
│    │    ├── kustomization.yaml
│    │    └── monitor.yaml
│    ├── rbac
│    │    ├── auth_proxy_client_clusterrole.yaml
│    │    ├── auth_proxy_role_binding.yaml
│    │    ├── auth_proxy_role.yaml
│    │    ├── auth_proxy_service.yaml
│    │    ├── kustomization.yaml
│    │    ├── leader_election_role_binding.yaml
│    │    ├── leader_election_role.yaml
│    │    ├── presentation_editor_role.yaml
│    │    ├── presentation_viewer_role.yaml
│    │    ├── role_binding.yaml
│    │    └── role.yaml
│    ├── samples
│    │    └── example_v1alpha1_presentation.yaml
│    └── webhook
│        ├── kustomization.yaml
│        ├── kustomizeconfig.yaml
│        └── service.yaml
├── controllers
│    ├── presentation_controller.go
│    └── suite_test.go
├── Dockerfile
├── go.mod
├── go.sum
├── hack
│    └── boilerplate.go.txt
├── main.go
├── Makefile
└── PROJECT
```

## Define API properties

Define api properties in `./api/v1alpha1/presentation_types.go`.

Note: After changing `./api/v1alpha1/presentation_types.go`, remenber to run `make manifests` again to update `./config/crd/bases/example.meetup.com_presentations.yaml`. 

## Develop controllers

Add your logic code in `./controllers/presentation_controller.go` to control CR(Custom resource).

## Test

First, install CRD on kubernetes cluster by:

```
make install
```

Now you can check if the CRD was successfully installed by:

```
kubectl get CRD 
```

Second, run controllers locally by:

```
make run
```

Third, create or delete CR to test whether controllers are working properly. For example:

```
kubectl apply -f ./config/samples/example_v1alpha1_presentation.yaml
kubectl delete -f ./config/samples/example_v1alpha1_presentation.yaml
```

If everything goes well, then we can BUILD & PUSH the controller image, and deploy the controller on kubernetes cluster.

## Build & push

Actually, I did not build & push the image to my dockerhub, and the Dockerfile need to be modifying refered to [this](https://github.com/HaroldMua/k8s-redevlelop/blob/master/operator-examples/demomicroservice/Dockerfile).
```
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```

## Deploy

```
 make deploy  IMG=<some-registry>/<project-name>:tag
```

## Uninstall CRD

```
make uninstall
```
