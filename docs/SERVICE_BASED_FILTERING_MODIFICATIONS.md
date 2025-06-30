# Service-Based Node Filtering Implementation Modifications

This document outlines the modifications required to implement service-based node filtering, where only nodes running pods for a specified service are registered to the load balancer.

## Files to Modify

### 1. `api/v1alpha1/lbregistrar_types.go`

#### Modification: Add FilterByEndpoints field to ServiceSpec

```go
// ServiceSpec defines the target service to get NodePort from.
type ServiceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	// Port is the port of the service.
	// It can be a port name or a port number.
	// +kubebuilder:validation:Required
	Port intstr.IntOrString `json:"port"`

	// FilterByEndpoints enables filtering nodes based on service endpoints.
	// When true, only nodes running pods for this service are registered to the load balancer.
	// +optional
	FilterByEndpoints bool `json:"filterByEndpoints,omitempty"`
}
```

### 2. `internal/controller/lbregistrar_controller.go`

#### Modification A: Add Endpoints RBAC permissions

```go
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
```

#### Modification B: Add Endpoints watching to SetupWithManager

```go
// SetupWithManager sets up the controller with the Manager.
func (r *LBRegistrarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.LBRegistrar{}).
		Watches(&corev1.Node{}, &NodeHandler{Client: r.Client, Recorder: r.Recorder},
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&corev1.Endpoints{}, &EndpointsHandler{Client: r.Client, Recorder: r.Recorder},
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
```

#### Modification C: Implement service-based filtering in register function

```go
// register handles backend registration for the given LBRegistrar resource.
func register(ctx context.Context, clnt client.Client, registrar *api.LBRegistrar) (configErr error, regErr error) {
	logger := log.FromContext(ctx)

	provider, configErr := getConfigurationProvider(ctx, clnt, registrar)
	if configErr != nil {
		return
	}

	spec := registrar.Spec
	if spec.Service != nil {
		logger.Info("service is specified, trying to get nodePort from service", "service", spec.Service.Name, "namespace", spec.Service.Namespace)
		nodePort, err := getNodePortFromService(ctx, clnt, spec.Service)
		if err != nil {
			regErr = fmt.Errorf("failed to get nodePort from service: %w", err)
			return
		}
		logger.Info("successfully got nodePort from service", "nodePort", nodePort)
		spec.NodePort = nodePort
	}

	var nodes *corev1.NodeList
	
	// Service-based filtering enabled
	if spec.Service != nil && spec.Service.FilterByEndpoints {
		logger.Info("service-based node filtering enabled")
		filteredNodes, err := getNodesForService(ctx, clnt, spec.Service)
		if err != nil {
			regErr = fmt.Errorf("failed to get nodes for service: %w", err)
			return
		}
		nodes = filteredNodes
		logger.Info("filtered nodes based on service", "nodeCount", len(nodes.Items))
	} else {
		// Use all nodes as before
		nodes = &corev1.NodeList{}
		configErr = clnt.List(ctx, nodes)
		if configErr != nil {
			configErr = client.IgnoreNotFound(configErr)
			return
		}
		logger.Info("using all nodes", "nodeCount", len(nodes.Items))
	}

	logger.V(1).Info("found node", "count", len(nodes.Items))
	logger.V(2).Info("found nodes", "nodes", nodes.Items)

	regErr = oci.RegisterBackends(ctx, provider, spec, nodes)
	return
}
```

#### Modification D: Add service-based node filtering function

```go
// getNodesForService retrieves nodes that are running pods for the specified service.
func getNodesForService(ctx context.Context, clnt client.Client, svcSpec *api.ServiceSpec) (*corev1.NodeList, error) {
	logger := log.FromContext(ctx)
	
	// Get Endpoints
	endpoints := &corev1.Endpoints{}
	endpointsKey := client.ObjectKey{
		Namespace: svcSpec.Namespace,
		Name:      svcSpec.Name,
	}
	if err := clnt.Get(ctx, endpointsKey, endpoints); err != nil {
		return nil, fmt.Errorf("failed to get endpoints for service %s/%s: %w", svcSpec.Namespace, svcSpec.Name, err)
	}

	// Collect Pod IP addresses from Endpoints
	podIPs := make(map[string]bool)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.IP != "" {
				podIPs[address.IP] = true
			}
		}
	}

	if len(podIPs) == 0 {
		logger.Info("no endpoints found for service", "service", svcSpec.Name)
		return &corev1.NodeList{}, nil
	}

	// Get all pods and create IP->Node name mapping
	pods := &corev1.PodList{}
	if err := clnt.List(ctx, pods, client.InNamespace(svcSpec.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", svcSpec.Namespace, err)
	}

	nodeNames := make(map[string]bool)
	for _, pod := range pods.Items {
		if podIPs[pod.Status.PodIP] && pod.Spec.NodeName != "" {
			nodeNames[pod.Spec.NodeName] = true
		}
	}

	// Create NodeList containing only relevant nodes
	allNodes := &corev1.NodeList{}
	if err := clnt.List(ctx, allNodes); err != nil {
		return nil, fmt.Errorf("failed to list all nodes: %w", err)
	}

	filteredNodes := &corev1.NodeList{}
	for _, node := range allNodes.Items {
		if nodeNames[node.Name] {
			filteredNodes.Items = append(filteredNodes.Items, node)
		}
	}

	logger.Info("filtered nodes for service", 
		"service", fmt.Sprintf("%s/%s", svcSpec.Namespace, svcSpec.Name),
		"totalPodIPs", len(podIPs),
		"uniqueNodes", len(nodeNames),
		"filteredNodes", len(filteredNodes.Items))

	return filteredNodes, nil
}
```

### 3. `internal/controller/endpoints_handler.go` (New File)

#### Content: Endpoints event handler

```go
/*
GNU GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (c) 2024-25 Norihiro Seto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

For the full license text, please visit: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

// EndpointsHandler handles Endpoints resource events for service-based node filtering.
type EndpointsHandler struct {
	client.Client
	Recorder record.EventRecorder
}

// Create handles endpoints creation events.
func (eh *EndpointsHandler) Create(ctx context.Context, evt event.TypedCreateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.Object.GetName(), "namespace", evt.Object.GetNamespace())
	logger.V(1).Info("endpoints creation")
	eh.handleEndpointsChange(ctx, evt.Object)
}

// Update handles endpoints update events.
func (eh *EndpointsHandler) Update(ctx context.Context, evt event.TypedUpdateEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.ObjectNew.GetName(), "namespace", evt.ObjectNew.GetNamespace())
	logger.V(1).Info("endpoints update")
	eh.handleEndpointsChange(ctx, evt.ObjectNew)
}

// Delete handles endpoints deletion events.
func (eh *EndpointsHandler) Delete(ctx context.Context, evt event.TypedDeleteEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	logger := log.FromContext(ctx, "endpoints", evt.Object.GetName(), "namespace", evt.Object.GetNamespace())
	logger.V(1).Info("endpoints deletion")
	eh.handleEndpointsChange(ctx, evt.Object)
}

func (eh *EndpointsHandler) Generic(ctx context.Context, evt event.TypedGenericEvent[client.Object], _ workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	// Do nothing
}

// handleEndpointsChange processes endpoints changes and updates affected LBRegistrar resources.
func (eh *EndpointsHandler) handleEndpointsChange(ctx context.Context, obj client.Object) {
	logger := log.FromContext(ctx, "endpoints", obj.GetName(), "namespace", obj.GetNamespace())
	
	// Find LBRegistrar resources that reference this service
	list := &api.LBRegistrarList{}
	if err := eh.List(ctx, list); err != nil {
		logger.Error(err, "failed to list LBRegistrar resources")
		return
	}

	affectedCount := 0
	for _, lb := range list.Items {
		// Check if this LBRegistrar uses service-based filtering for this service
		if lb.Spec.Service != nil && 
		   lb.Spec.Service.FilterByEndpoints && 
		   lb.Spec.Service.Namespace == obj.GetNamespace() && 
		   lb.Spec.Service.Name == obj.GetName() {
			
			// Update status to trigger reconciliation
			if lb.Status.Phase != api.PhasePending && lb.Status.Phase != api.PhaseNew {
				lb.Status.Phase = api.PhasePending
				if err := eh.Status().Update(ctx, &lb); err != nil {
					logger.Error(err, "failed to update LBRegistrar status", "registrar", lb.Name)
					continue
				}
				eh.Recorder.Event(&lb, corev1.EventTypeNormal, "EndpointsChanged", 
					fmt.Sprintf("Service %s/%s endpoints changed, triggering reconciliation", 
						obj.GetNamespace(), obj.GetName()))
				affectedCount++
			}
		}
	}
	
	if affectedCount > 0 {
		logger.Info("triggered reconciliation for affected LBRegistrar resources", "count", affectedCount)
	}
}
```

## Post-Implementation Workflow

1. **LBRegistrar Creation**: 
   - When `Service.FilterByEndpoints=true`, check service endpoints
   - Register only nodes running pods for the service to the load balancer

2. **Pod Changes**:
   - Endpoints resource is automatically updated
   - `EndpointsHandler` detects changes
   - Affected LBRegistrars are set to `PhasePending`
   - Re-reconciliation updates load balancer configuration

3. **Configuration Example**:
   ```yaml
   apiVersion: nodes.peppy-ratio.dev/v1alpha1
   kind: LBRegistrar
   spec:
     loadBalancerId: "ocid1.loadbalancer.oc1..."
     service:
       name: "my-web-app"
       namespace: "production"
       port: 80
       filterByEndpoints: true  # Enable service-based filtering
     backendSetName: "web-backends"
   ```

## Required Commands

After implementation, run the following commands:

```bash
# Regenerate CRD schemas
make generate
make manifests

# Run tests
make test

# Code quality checks
make fmt
make vet
make lint
```

This implementation enables registering only nodes running the specified service's pods to the load balancer, with dynamic updates when pod changes occur.