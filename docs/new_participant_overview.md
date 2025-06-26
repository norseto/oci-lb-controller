Below is an overview of the repository’s purpose, major directories and files, and suggestions for where to learn more.

---

## Purpose of the Project
The repository implements **Oracle Cloud LoadBalancer Registrar**, a Kubernetes operator written in Go.  
It automatically registers Kubernetes nodes with Oracle Cloud LoadBalancers and supports both classic and network load balancers. Key features include automatic node registration, dynamic management of backends, service integration to detect NodePorts, and support for multiple `LBRegistrar` resources.

---

## Repository Layout
```
.
├── api/                # Custom resource definitions (LBRegistrar types)
├── cmd/                # Entry point (main.go) setting up the controller manager
├── internal/
│   └── controller/     # Reconciler, node handler, OCI provider, models
│       └── cloud/oci/  # OCI load balancer interactions
├── config/             # Kustomize manifests for CRDs and deployment
├── test/               # e2e tests and utilities
├── Dockerfile          # Build container image
├── Makefile            # Build/test/deploy targets
└── dist/               # Pre-built install YAML
```

### Core Components
- **API Types (`api/v1alpha1/`)**  
  Define the `LBRegistrar` custom resource with fields such as LoadBalancerId, BackendSetName and API key parameters.
- **Controller (`internal/controller/`)**  
  `LBRegistrarReconciler` reconciles both `LBRegistrar` objects and Kubernetes Node events. `node_handler.go` watches node additions/deletions to trigger re-registration.
- **Cloud Providers (`internal/controller/cloud/oci/`)**  
  Implements OCI-specific load balancer logic for both classic and network load balancers.

### Configuration
Deployment is managed with Kustomize. Example CRDs and sample resources live under `config/`, and a prebuilt installation manifest is generated in `dist/install.yaml`.

---

## Development Workflow
Common make targets are provided:

- Build and run tests:  
  `make build`, `make test`, `make test-e2e`
- Linting and formatting:  
  `make fmt`, `make vet`, `make lint`, `make lint-fix`
- Deployment helpers:  
  `make install`, `make deploy`, `make undeploy`, `make build-installer`

Tests use the Ginkgo/Gomega framework, with unit tests under `internal/controller` and end-to-end tests in `test/e2e/`.

---

## Suggested Next Steps

1. **Understand the CRD**  
   Review `api/v1alpha1/lbregistrar_types.go` to see all fields of the `LBRegistrar` resource.

2. **Explore the Controller Logic**  
   Examine `internal/controller/lbregistrar_controller.go` and `internal/controller/node_handler.go` for the reconciliation flow and node event handling.

3. **Check OCI Integration**  
   Dive into `internal/controller/cloud/oci/provider.go` and the `loadbalancer` and `networkloadbalancer` directories to see how backends are registered with OCI.

4. **Run and Test Locally**  
   Use the Makefile commands to build (`make build`), run tests (`make test`), and spin up end-to-end tests (`make test-e2e`) against a local Kind cluster.

5. **Deploy on a Cluster**  
   After building an image (`make docker-build`), install the CRDs (`make install`) and deploy (`make deploy`). Sample manifests in `config/samples` can help create initial `LBRegistrar` resources.

6. **Examine Generated Artifacts**  
   The `dist/install.yaml` file contains the consolidated CRDs and deployment YAML, useful for manual installations or debugging.

---

This summary should help you start navigating the project, understand where key functionality resides, and know how to build, test, and deploy the operator. For deeper details, refer to the code comments in each package and the Makefile targets.
