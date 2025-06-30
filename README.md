# Oracle Cloud LoadBalancer Registrar

Kubernetes Operator for Oracle Cloud Infrastructure (OCI) that automatically registers Kubernetes worker nodes to Oracle Cloud LoadBalancers. This operator monitors Kubernetes nodes and dynamically manages their registration with OCI Load Balancers (both Classic and Network Load Balancers).

## Overview

This operator provides the following features:

- **Automatic node registration**: Automatically registers newly added nodes in the Kubernetes cluster to the backend set of an Oracle Cloud Infrastructure (OCI) Load Balancer.
- **Dynamic management**: Continuously monitors node additions and removals and updates the Load Balancer configuration accordingly.
- **Service integration**: Automatically detects the NodePort of a Kubernetes Service resource and uses it to register the node with the Load Balancer.
- **Multi-service support**: Single `LBRegistrar` resource can manage multiple services and backend sets, eliminating conflicts.
- **Multi-Load Balancer support**: Supports multiple `LBRegistrar` resources to register nodes with different Load Balancers.

## Usage

1. **Setup OCI LoadBalancer**:
    Create a LoadBalancer in OCI Console or using OCI CLI.
    Make sure to create a backend set for the LoadBalancer.

2. **Create NodePort Service**:
    Create a NodePort service in Kubernetes to expose the application.
    
3. **Create OCI user and API key**:
    Create a user and API key in OCI Console or using OCI CLI.
    This user will be used to authenticate to update the LoadBalancer.
    
4. **OCI API Key secret creation**:
    ```bash
    kubectl -n key-namespace create secret generic oci-api-key \
    --from-file=private-key=path/to/your/oci_api_key.pem
    ```

5. **Deploy Operator**:
   ```bash
   kubectl apply -f dist/install.yaml
   ```

6. **Create LBRegistrar resource**:

    **Single Service (Legacy):**
    ```yaml
    apiVersion: nodes.peppy-ratio.dev/v1alpha1
    kind: LBRegistrar
    metadata:
      name: my-app-registrar
    spec:
      loadBalancerId: "ocid1.loadbalancer.oc1.ap-tokyo-1.xxxxx"
      backendSetName: "my-backend-set"
      weight: 1
      apiKey:
        user: "ocid1.user.oc1..xxxxx"
        fingerprint: "aa:bb:cc:dd:ee"
        tenancy: "ocid1.tenancy.oc1..xxxxx"
        region: "ap-tokyo-1"
        privateKey:
          namespace: key-namespace
          secretKeyRef:
            name: oci-api-key
            key: private-key
      service:
        name: "my-nodeport-service"
        namespace: "default"
        port: http2
        filterByEndpoints: true
    ```

    **Multiple Services (Recommended):**
    ```yaml
    apiVersion: nodes.peppy-ratio.dev/v1alpha1
    kind: LBRegistrar
    metadata:
      name: my-app-multiservice-registrar
    spec:
      loadBalancerId: "ocid1.loadbalancer.oc1.ap-tokyo-1.xxxxx"
      apiKey:
        user: "ocid1.user.oc1..xxxxx"
        fingerprint: "aa:bb:cc:dd:ee"
        tenancy: "ocid1.tenancy.oc1..xxxxx"
        region: "ap-tokyo-1"
        privateKey:
          namespace: key-namespace
          secretKeyRef:
            name: oci-api-key
            key: private-key
      services:
        - name: "my-nodeport-service"
          namespace: "default"
          port: https
          backendSetName: "https-backend"
          weight: 1
          filterByEndpoints: true
        - name: "my-nodeport-service"
          namespace: "default"
          port: http2
          backendSetName: "http-backend"
          weight: 1
          filterByEndpoints: true
    ```

## Working Principle

1. When `LBRegistrar` resource is created, the operator establishes a connection to the OCI LoadBalancer
2. The operator monitors all nodes in the Kubernetes cluster
3. When a new node is added, it automatically registers the node with the LoadBalancer using the specified NodePort
4. When a node is removed, it automatically removes the node from the LoadBalancer
5. If Service resources are specified, the operator dynamically retrieves the NodePort from each Service
6. For multi-service configurations, all services are processed sequentially within a single reconciliation cycle to avoid conflicts
7. **WorkRequest Management**: Each backend update waits for OCI WorkRequest completion to ensure no state conflicts

## Multi-Service Benefits

The new multi-service support eliminates the OCI Load Balancer conflict issues that occurred when multiple `LBRegistrar` resources targeted the same load balancer:

- **Conflict Resolution**: No more "Invalid State Transition" errors from concurrent updates
- **Simplified Management**: Single resource manages multiple backend sets
- **Atomic Operations**: All backend updates happen within one reconciliation cycle
- **WorkRequest Synchronization**: Waits for each OCI operation to complete before proceeding to the next
- **Backward Compatibility**: Existing single-service configurations continue to work unchanged

## Reliability Features

### WorkRequest Management
The controller implements robust WorkRequest handling for OCI operations:

- **Completion Waiting**: Each `UpdateBackendSet` operation waits for WorkRequest completion (up to 5 minutes)
- **Status Monitoring**: Polls WorkRequest status every 5 seconds with detailed logging
- **Error Handling**: Properly handles Failed, Canceled, and timeout scenarios
- **Network Load Balancer Support**: Full WorkRequest implementation for NLB operations

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
