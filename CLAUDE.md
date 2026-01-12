# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the Oracle Cloud LoadBalancer Registrar, a Kubernetes Operator built with Kubebuilder framework that automatically registers new Kubernetes nodes to Oracle Cloud LoadBalancers. It's written in Go and follows the controller-runtime pattern.

## Development Commands

### Building and Testing
- `make build` - Build the manager binary
- `make test` - Run unit tests (excludes e2e tests)
- `make test-e2e` - Run end-to-end tests against a Kind cluster
- `make run` - Run the controller locally against your configured cluster

### Code Quality
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint
- `make lint-fix` - Run golangci-lint with automatic fixes

### Code Generation
- `make generate` - Generate DeepCopy methods for API types
- `make manifests` - Generate CRDs, RBAC, and webhook configurations

### Docker Operations
- `make docker-build` - Build Docker image (default: controller:latest)
- `make docker-push` - Push Docker image
- `make docker-buildx` - Build multi-platform images

### Deployment
- `make install` - Install CRDs to cluster
- `make deploy` - Deploy controller to cluster
- `make undeploy` - Remove controller from cluster
- `make build-installer` - Generate consolidated YAML in dist/install.yaml

## Architecture

### Core Components

**API Types** (`api/v1alpha1/`):
- `LBRegistrar` - Custom Resource that defines load balancer registration configuration
- Key fields: LoadBalancerId, Services (or legacy Service), Weight, BackendSetName
- Multi-service support: Services array allows multiple service/backend set configurations in single resource

**Controller** (`internal/controller/`):
- `LBRegistrarReconciler` - Main controller that watches Node and LBRegistrar resources
- `node_handler.go` - Handles Node events and manages registrations
- `endpoints_handler.go` - Handles Endpoints events for service-based filtering (supports multi-service)
- Reconciles both Node changes and LBRegistrar spec changes
- Multi-service support: `registerMultipleServices()` function processes all services sequentially

**Cloud Providers** (`internal/controller/cloud/oci/`):
- `provider.go` - OCI configuration and authentication
- `loadbalancer/` - OCI Load Balancer operations
- `networkloadbalancer/` - Network Load Balancer operations with WorkRequest management

**Models** (`internal/controller/models/`):
- Common data structures shared across components

### Authentication
Uses OCI API keys with the following configuration:
- Tenancy, User, Region, Fingerprint from LBRegistrar spec
- Private key from Kubernetes Secret

### Deployment Structure
- Uses Kustomize for configuration management
- RBAC configured for node and secret access
- Prometheus metrics and health checks enabled
- Leader election for high availability

## Testing

The project uses Ginkgo/Gomega for testing:
- Unit tests in `*_test.go` files
- E2E tests in `test/e2e/`
- Test utilities in `test/utils/`
- Coverrage: 80%

Run specific tests with:
- `go test ./internal/controller/` - Run controller tests only
- `go test ./internal/controller/cloud/oci/` - Run OCI provider tests only

## Configuration

Key configuration files:
- `config/` - Kustomize manifests for deployment
- `config/samples/` - Example LBRegistrar resources
- `config/test/` - Test configurations including API key examples

## Controller Architecture Details

The controller follows a dual-reconciliation pattern:
- **LBRegistrar reconciliation**: Triggered by changes to LBRegistrar resources
- **Node reconciliation**: Triggered by Node events (add/remove/update)
- **Endpoints reconciliation**: Triggered by Endpoints changes (for service-based filtering)

Both reconciliation paths converge in the main controller which manages the actual OCI LoadBalancer operations. The controller maintains state consistency by:
1. Watching LBRegistrar, Node, and Endpoints resources
2. Cross-referencing existing backend registrations with current cluster state
3. Performing differential updates (add/remove only changed nodes)
4. **Multi-service support**: Processing all services in a single reconciliation cycle to prevent conflicts

## Development Notes

- The controller uses leader election for high availability in multi-replica deployments
- OCI authentication is handled through API keys stored in Kubernetes secrets
- Both OCI Load Balancers and Network Load Balancers are supported via separate provider implementations
- The project includes comprehensive RBAC configurations for production deployment
- **WorkRequest Management**: Network Load Balancer operations include WorkRequest completion waiting to prevent state conflicts
