# Multi-Service Support Design for LBRegistrar

## Overview

This document outlines the design for extending the LBRegistrar custom resource to support multiple services and ports within a single resource definition, addressing the current limitation that causes conflicts when multiple LBRegistrar resources target the same Oracle Cloud Load Balancer.

## Problem Statement

### Current Issue
The existing architecture requires separate LBRegistrar resources for each service/port combination targeting the same load balancer. This creates race conditions and conflicts when multiple controllers attempt to update the same OCI Load Balancer simultaneously, resulting in errors like:

```
Error Code: Conflict. Invalid State Transition of NLB lifeCycle state from Updating to Updating
```

### Root Cause Analysis
1. **Concurrent Updates**: Multiple LBRegistrar resources targeting the same load balancer create simultaneous update operations
2. **Lack of Coordination**: No synchronization mechanism between different LBRegistrar controllers
3. **OCI API Limitations**: Oracle Cloud Load Balancers cannot handle concurrent state transitions

## Proposed Solution

### Multi-Service Spec Design

Extend the `LBRegistrarSpec` to support multiple services within a single resource:

```go
type LBRegistrarSpec struct {
    LoadBalancerId string `json:"loadBalancerId"`
    BackendSetName string `json:"backendSetName"`
    ApiKey         ApiKeySpec `json:"apiKey"`
    
    // New field: Support multiple services/ports
    Services []ServiceSpec `json:"services,omitempty"`
    
    // Deprecated fields (maintained for backward compatibility)
    Service  *ServiceSpec `json:"service,omitempty"`
    NodePort int          `json:"nodePort,omitempty"`
    Weight   int          `json:"weight,omitempty"`
}

type ServiceSpec struct {
    Name              string             `json:"name"`
    Namespace         string             `json:"namespace"`
    Port              intstr.IntOrString `json:"port"`
    FilterByEndpoints bool               `json:"filterByEndpoints,omitempty"`
    Weight            int                `json:"weight,omitempty"`
    
    // New optional field for backend set name override
    BackendSetName    string             `json:"backendSetName,omitempty"`
}
```

## Implementation Plan

### Phase 1: API Extension
1. **Update API Types** (`api/v1alpha1/lbregistrar_types.go`)
   - Add `Services []ServiceSpec` field
   - Add `BackendSetName` to `ServiceSpec` for per-service backend set override
   - Add `Weight` to `ServiceSpec` for per-service weight configuration
   - Mark existing single-service fields as deprecated

2. **Generate CRD Updates**
   - Run `make generate` and `make manifests`
   - Update validation rules for the new multi-service structure

### Phase 2: Controller Logic Updates
1. **Update LBRegistrar Controller** (`internal/controller/lbregistrar_controller.go`)
   - Modify reconciliation logic to handle multiple services
   - Implement backward compatibility with existing single-service specs
   - Add validation to prevent mixing old and new service definitions

2. **Update Node Handler** (`internal/controller/node_handler.go`)
   - Extend node registration logic to handle multiple service endpoints
   - Implement service-specific endpoint filtering
   - Handle per-service backend set operations

3. **Update Cloud Provider Logic**
   - Modify OCI Load Balancer operations to batch multiple backend updates
   - Implement transaction-like operations for multi-service updates
   - Add rollback capability for failed multi-service operations

### Phase 3: Testing and Validation
1. **Unit Tests**
   - Test backward compatibility with existing single-service configurations
   - Test multi-service configurations
   - Test validation logic for invalid configurations

2. **Integration Tests**
   - Test multi-service registration with actual OCI Load Balancers
   - Test conflict resolution and error handling
   - Test migration from single-service to multi-service configurations

### Phase 4: Documentation and Migration
1. **Update Documentation**
   - Update main README.md with multi-service examples and usage
   - Create migration guide for existing deployments
   - Update example configurations in config/samples/
   - Update CLAUDE.md with new multi-service architecture details
   - Document new multi-service capabilities

2. **Migration Tools**
   - Create utility scripts to convert existing single-service LBRegistrars to multi-service format
   - Provide validation tools for new configurations

## Configuration Examples

### Current Configuration (Multiple Resources)
```yaml
# Current problematic approach
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-https
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxx
  backendSetName: ingress
  service:
    name: istio-ingressgateway
    namespace: istio-system
    port: https
  # ... other fields
---
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-http
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxx  # Same LB!
  backendSetName: ingress-http
  service:
    name: istio-ingressgateway
    namespace: istio-system
    port: http2
  # ... other fields
```

### Proposed Configuration (Single Resource)
```yaml
# New consolidated approach
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-gateway
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxx
  apiKey:
    # ... API key configuration
  services:
    - name: istio-ingressgateway
      namespace: istio-system
      port: https
      backendSetName: ingress
      weight: 1
      filterByEndpoints: true
    - name: istio-ingressgateway
      namespace: istio-system
      port: http2
      backendSetName: ingress-http
      weight: 1
      filterByEndpoints: true
```

## Benefits

### Immediate Benefits
1. **Conflict Resolution**: Eliminates race conditions between multiple LBRegistrar resources
2. **Simplified Management**: Single resource to manage multiple service endpoints
3. **Atomic Operations**: All backend set updates occur within a single reconciliation cycle

### Long-term Benefits
1. **Better Resource Utilization**: Reduced controller overhead with fewer resources to watch
2. **Improved Observability**: Centralized status reporting for all services on a load balancer
3. **Enhanced Scalability**: Better performance with fewer reconciliation loops

## Backward Compatibility

The implementation will maintain full backward compatibility:
- Existing single-service configurations will continue to work unchanged
- Controllers will automatically detect and handle both old and new configuration formats
- Migration can be performed gradually without service interruption

## Risk Assessment

### Low Risk
- API extension using optional fields maintains backward compatibility
- Gradual migration approach minimizes service disruption

### Medium Risk
- Controller logic complexity increases with dual-mode support
- Testing requirements increase to cover both configuration types

### Mitigation Strategies
- Comprehensive test coverage for both configuration modes
- Phased rollout with extensive validation
- Clear migration documentation and tooling

## Timeline

- **Week 1-2**: API extension and CRD generation
- **Week 3-4**: Controller logic implementation
- **Week 5-6**: Testing and validation
- **Week 7**: Documentation updates (README.md, CLAUDE.md, samples) and migration tools
- **Week 8**: Release and deployment

## Success Criteria

1. **Functionality**: Multi-service LBRegistrar resources successfully manage multiple backend sets
2. **Compatibility**: Existing single-service configurations continue to work without modification
3. **Reliability**: No conflicts or race conditions when using multi-service configurations
4. **Performance**: Improved reconciliation performance with consolidated resources
5. **WorkRequest Management**: Proper OCI WorkRequest handling ensures operation completion and prevents state conflicts

## Implementation Status

### ✅ Completed Features

- **Multi-service API**: `Services []ServiceSpec` field added to LBRegistrarSpec
- **Backward Compatibility**: Existing single-service configurations continue to work
- **Controller Logic**: `registerMultipleServices()` function processes multiple services sequentially
- **Endpoints Handler**: Multi-service support for service-based filtering
- **WorkRequest Management**: Network Load Balancer operations wait for WorkRequest completion
- **Documentation**: Updated README, CLAUDE.md, and sample configurations

### 🔧 Technical Implementation

**WorkRequest Handling:**
- `waitForWorkRequestCompletion()` function polls WorkRequest status every 5 seconds
- Maximum 5-minute timeout with proper error handling
- Detailed logging for operational visibility
- Currently implemented for Network Load Balancer operations

**Conflict Resolution:**
- Sequential processing of services within single reconciliation cycle
- WorkRequest completion waiting prevents OCI state transition conflicts
- No more "Invalid State Transition" errors during concurrent updates