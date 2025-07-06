# Multi-Service Migration Guide

This guide helps you migrate from the legacy single-service LBRegistrar configuration to the new multi-service configuration that eliminates OCI Load Balancer conflicts.

## Background

Prior to version 0.5.0-alpha.4, managing multiple services for the same OCI Load Balancer required separate LBRegistrar resources. This approach caused conflicts when multiple controllers attempted to update the same load balancer simultaneously, resulting in errors like:

```
Error Code: Conflict. Invalid State Transition of NLB lifeCycle state from Updating to Updating
```

The new multi-service support resolves these conflicts by consolidating all services for a load balancer into a single resource.

## Migration Steps

### Step 1: Identify Conflicting Resources

Find all LBRegistrar resources that target the same `loadBalancerId`:

```bash
kubectl get lbregistrar -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.loadBalancerId}{"\n"}{end}' | sort -k2
```

### Step 2: Create Multi-Service Configuration

For each group of resources sharing the same `loadBalancerId`, create a new multi-service LBRegistrar:

**Before (Multiple Resources):**
```yaml
# Resource 1
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-https
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxxxx
  backendSetName: ingress
  service:
    name: istio-ingressgateway
    namespace: istio-system
    port: https
    filterByEndpoints: true
  # ... API key config
---
# Resource 2  
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-http
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxxxx  # Same LB!
  backendSetName: ingress-http
  service:
    name: istio-ingressgateway
    namespace: istio-system
    port: http2
    filterByEndpoints: true
  # ... API key config
```

**After (Single Resource):**
```yaml
apiVersion: nodes.peppy-ratio.dev/v1alpha1
kind: LBRegistrar
metadata:
  name: lbregistrar-ingress-consolidated
spec:
  loadBalancerId: ocid1.loadbalancer.oc1.ap-tokyo-1.xxxxx
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

### Step 3: Apply Migration

1. **Create the new multi-service resource:**
   ```bash
   kubectl apply -f new-multiservice-lbregistrar.yaml
   ```

2. **Wait for the new resource to be ready:**
   ```bash
   kubectl get lbregistrar lbregistrar-ingress-consolidated -w
   ```

3. **Delete the old conflicting resources:**
   ```bash
   kubectl delete lbregistrar lbregistrar-ingress-https lbregistrar-ingress-http
   ```

### Step 4: Verify Migration

Check that the consolidated resource is managing all services correctly:

```bash
# Check resource status
kubectl get lbregistrar lbregistrar-ingress-consolidated -o yaml

# Check controller logs
kubectl logs -n oci-lb-controller-system deployment/oci-lb-controller-controller-manager -f
```

## Configuration Reference

### Multi-Service Field Mapping

| Legacy Field | Multi-Service Field | Notes |
|-------------|-------------------|-------|
| `spec.service` | `spec.services[].name/namespace/port` | Each service becomes an array element |
| `spec.backendSetName` | `spec.services[].backendSetName` | Can be different per service |
| `spec.weight` | `spec.services[].weight` | Can be different per service |
| N/A | `spec.services[].filterByEndpoints` | Copied from legacy service config |

### Field Priority

When using multi-service configuration:
- `spec.services` takes precedence over `spec.service` 
- `spec.services[].backendSetName` overrides `spec.backendSetName` for that service
- `spec.services[].weight` overrides `spec.weight` for that service

### Backward Compatibility

The new multi-service support is fully backward compatible:
- Existing single-service configurations continue to work
- No immediate migration required
- Controllers automatically detect configuration type

## Best Practices

1. **Group by Load Balancer**: Create one multi-service LBRegistrar per unique `loadBalancerId`

2. **Meaningful Names**: Use descriptive names that indicate the load balancer and services managed:
   ```yaml
   metadata:
     name: nlb-ingress-gateway-services  # Clear, descriptive name
   ```

3. **Service-specific Backend Sets**: Use different backend sets for different protocols/ports:
   ```yaml
   services:
     - port: https
       backendSetName: tls-backends
     - port: http2  
       backendSetName: http-backends
   ```

4. **Consistent Filtering**: Use the same `filterByEndpoints` setting for services from the same deployment:
   ```yaml
   services:
     - name: istio-ingressgateway
       filterByEndpoints: true
     - name: istio-ingressgateway  
       filterByEndpoints: true  # Consistent
   ```

## Troubleshooting

### Migration Issues

**Problem**: New multi-service resource stuck in PENDING
- **Solution**: Check that all referenced services exist and have NodePort type

**Problem**: Old resources still conflicting during migration
- **Solution**: Delete old resources immediately after creating new one

**Problem**: Backend sets not updating
- **Solution**: Verify backend set names exist in OCI Load Balancer

### Post-Migration Validation

1. **Check all services are processed:**
   ```bash
   kubectl logs -n oci-lb-controller-system deployment/oci-lb-controller-controller-manager | grep "processing service"
   ```

2. **Verify no conflicts in OCI:**
   - Monitor OCI Console for load balancer update status
   - Ensure no "Invalid State Transition" errors in logs

3. **Test application connectivity:**
   - Verify all service endpoints are reachable through load balancer
   - Test both HTTP and HTTPS if applicable

## Migration Script

Here's a helper script to automate the migration process:

```bash
#!/bin/bash
# migration-helper.sh

LB_ID="$1"
if [ -z "$LB_ID" ]; then
    echo "Usage: $0 <loadBalancerId>"
    exit 1
fi

echo "Finding LBRegistrar resources for Load Balancer: $LB_ID"

# Get all resources for this LB
RESOURCES=$(kubectl get lbregistrar -o json | jq -r --arg lb "$LB_ID" '.items[] | select(.spec.loadBalancerId == $lb) | .metadata.name')

if [ -z "$RESOURCES" ]; then
    echo "No LBRegistrar resources found for Load Balancer: $LB_ID"
    exit 1
fi

echo "Found resources:"
echo "$RESOURCES"

echo ""
echo "Generate consolidated configuration with:"
echo "kubectl get lbregistrar $RESOURCES -o yaml > migration-input.yaml"
echo ""
echo "Then manually create multi-service configuration and apply:"
echo "kubectl apply -f consolidated-lbregistrar.yaml"
echo ""
echo "Finally delete old resources:"
echo "kubectl delete lbregistrar $RESOURCES"
```

## Support

If you encounter issues during migration:

1. Check the [troubleshooting guide](troubleshooting.md)
2. Review controller logs for detailed error messages
3. Consult the [design document](multi-service-support-design.md) for technical details
4. Open an issue on GitHub with migration logs and configuration details
