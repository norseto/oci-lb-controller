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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/oracle/oci-go-sdk/v65/common"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/cloud/oci"
)

// LBRegistrarReconciler reconciles a LBRegistrar object
type LBRegistrarReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nodes.peppy-ratio.dev,resources=lbregistrars/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LBRegistrar object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *LBRegistrarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result := ctrl.Result{}
	shouldUpdate := false

	registrar := &api.LBRegistrar{}
	if err := r.Get(ctx, req.NamespacedName, registrar); err != nil {
		log.FromContext(ctx).Error(err, "unable to fetch LBRegistrar")
		return result, client.IgnoreNotFound(err)
	}

	logger := log.FromContext(ctx,
		"lbId", registrar.Spec.LoadBalancerId,
		"backendset", registrar.Spec.BackendSetName)
	ctx = log.IntoContext(ctx, logger)

	defer func() {
		if !shouldUpdate {
			return
		}
		if err := r.Status().Update(ctx, registrar); err != nil {
			logger.Error(err, "unable to update LBRegistrar status")
		}
	}()

	switch registrar.Status.Phase {
	case api.PhaseNew:
		logger.Info("reconciling pending registrar")
		registrar.Status.Phase = api.PhasePending
		shouldUpdate = true
	case api.PhasePending:
		logger.Info("reconciling pending registrar")
		provider, err := getConfigurationProvider(ctx, r.Client, registrar)
		if err != nil {
			logger.Error(err, "unable to create configuration provider")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to create configuration provider: %v", err)
			return result, fmt.Errorf("unable to create configuration provider: %w", err)
		}

		backends, err := oci.GetBackendSet(ctx, provider, registrar.Spec)
		if err != nil {
			logger.Error(err, "unable to get backend set")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to get backend set: %v", err)
			return result, err
		}
		logger.Info("got current backends", "backends", backends)
		registrar.Status.Phase = api.PhaseRegistering
		shouldUpdate = true
	case api.PhaseRegistering:
		logger.Info("reconciling registering registrar")
		confErr, regErr := register(ctx, r.Client, registrar)
		if confErr != nil {
			logger.Error(confErr, "unable to create configuration provider")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to create configuration provider: %v", confErr)
			registrar.Status.Phase = api.PhasePending
			shouldUpdate = true
			return result, confErr
		} else if regErr != nil {
			logger.Error(regErr, "unable to register backends")
			r.Recorder.Eventf(registrar, corev1.EventTypeWarning, "Failed", "unable to register backends: %v", regErr)
			result.RequeueAfter = 90 * time.Second
			return result, regErr
		}
		registrar.Status.Phase = api.PhaseReady
		shouldUpdate = true
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LBRegistrarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.LBRegistrar{}).
		Watches(&corev1.Node{}, &NodeHandler{Client: r.Client, Recorder: r.Recorder},
													builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&corev1.Endpoints{}, &EndpointsHandler{Client: r.Client, Recorder: r.Recorder}, //nolint:staticcheck
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

// getConfigurationProvider retrieves the OCI ConfigurationProvider using the API key and private key stored in a Kubernetes Secret.
func getConfigurationProvider(ctx context.Context, client client.Client, registrar *api.LBRegistrar) (common.ConfigurationProvider, error) {
	secSpec := registrar.Spec.ApiKey.PrivateKey
	privateKey, err := GetSecretValue(ctx, client, secSpec.Namespace, &secSpec.SecretKeyRef)
	if err != nil {
		return nil, err
	}

	provider, err := oci.NewConfigurationProvider(ctx, &registrar.Spec.ApiKey, privateKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// register handles backend registration for the given LBRegistrar resource.
func register(ctx context.Context, clnt client.Client, registrar *api.LBRegistrar) (configErr error, regErr error) {
	logger := log.FromContext(ctx)

	provider, configErr := getConfigurationProvider(ctx, clnt, registrar)
	if configErr != nil {
		return configErr, regErr
	}

	// Handle multi-service registration
	if len(registrar.Spec.Services) > 0 {
		logger.Info("multiple services specified, processing each service", "serviceCount", len(registrar.Spec.Services))
		regErr = registerMultipleServices(ctx, clnt, provider, registrar)
		return configErr, regErr
	}

	// Backward compatibility: single service registration
	spec := registrar.Spec
	if spec.Service != nil {
		logger.Info("single service specified, trying to get nodePort from service", "service", spec.Service.Name, "namespace", spec.Service.Namespace)
		nodePort, err := getNodePortFromService(ctx, clnt, spec.Service)
		if err != nil {
			regErr = fmt.Errorf("failed to get nodePort from service: %w", err)
			return configErr, regErr
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
			return configErr, regErr
		}
		nodes = filteredNodes
		logger.Info("filtered nodes based on service", "nodeCount", len(nodes.Items))
	} else {
		// Use all nodes as before
		nodes = &corev1.NodeList{}
		configErr = clnt.List(ctx, nodes)
		if configErr != nil {
			configErr = client.IgnoreNotFound(configErr)
			return configErr, regErr
		}
		logger.Info("using all nodes", "nodeCount", len(nodes.Items))
	}

	logger.V(1).Info("found node", "count", len(nodes.Items))
	logger.V(2).Info("found nodes", "nodes", nodes.Items)

	regErr = oci.RegisterBackends(ctx, provider, spec, nodes)
	return configErr, regErr
}

// getNodePortFromService retrieves the NodePort value from a Kubernetes Service specified by svcSpec.
func getNodePortFromService(ctx context.Context, clnt client.Client, svcSpec *api.ServiceSpec) (int, error) {
	logger := log.FromContext(ctx)

	svc := &corev1.Service{}
	svcKey := client.ObjectKey{
		Namespace: svcSpec.Namespace,
		Name:      svcSpec.Name,
	}
	if err := clnt.Get(ctx, svcKey, svc); err != nil {
		return 0, fmt.Errorf("failed to get service %s/%s: %w", svcSpec.Namespace, svcSpec.Name, err)
	}

	if svc.Spec.Type != corev1.ServiceTypeNodePort {
		return 0, fmt.Errorf("service %s/%s is not of type NodePort", svc.Namespace, svc.Name)
	}

	for _, port := range svc.Spec.Ports {
		switch svcSpec.Port.Type {
		case intstr.Int:
			if port.Port == svcSpec.Port.IntVal {
				if port.NodePort == 0 {
					return 0, fmt.Errorf("nodePort is not allocated for port %d in service %s/%s", port.Port, svc.Namespace, svc.Name)
				}
				logger.Info("found matching port by number", "port", port.Port, "nodePort", port.NodePort)
				return int(port.NodePort), nil
			}
		case intstr.String:
			if port.Name == svcSpec.Port.StrVal {
				if port.NodePort == 0 {
					return 0, fmt.Errorf("nodePort is not allocated for port %s in service %s/%s", port.Name, svc.Namespace, svc.Name)
				}
				logger.Info("found matching port by name", "portName", port.Name, "nodePort", port.NodePort)
				return int(port.NodePort), nil
			}
		}
	}

	return 0, fmt.Errorf("no matching port found for %v in service %s/%s", svcSpec.Port, svc.Namespace, svc.Name)
}

// getNodesForService retrieves nodes that are running pods for the specified service.
func getNodesForService(ctx context.Context, clnt client.Client, svcSpec *api.ServiceSpec) (*corev1.NodeList, error) {
	logger := log.FromContext(ctx)

	// Get Endpoints
	endpoints := &corev1.Endpoints{} //nolint:staticcheck
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

// registerMultipleServices handles backend registration for multiple services in a single LBRegistrar resource.
func registerMultipleServices(ctx context.Context, clnt client.Client, provider common.ConfigurationProvider, registrar *api.LBRegistrar) error {
	logger := log.FromContext(ctx)

	// Process each service and collect all backend registrations
	for i, svcSpec := range registrar.Spec.Services {
		logger.Info("processing service", "index", i, "service", svcSpec.Name, "namespace", svcSpec.Namespace)

		// Create a spec for this specific service
		serviceSpec := registrar.Spec
		serviceSpec.Service = &svcSpec

		// Override backend set name if specified in service
		if svcSpec.BackendSetName != "" {
			serviceSpec.BackendSetName = svcSpec.BackendSetName
		}

		// Override weight if specified in service
		if svcSpec.Weight > 0 {
			serviceSpec.Weight = svcSpec.Weight
		}

		// Get nodePort from service
		nodePort, err := getNodePortFromService(ctx, clnt, &svcSpec)
		if err != nil {
			return fmt.Errorf("failed to get nodePort from service %s/%s: %w", svcSpec.Namespace, svcSpec.Name, err)
		}
		serviceSpec.NodePort = nodePort
		logger.Info("got nodePort for service", "service", svcSpec.Name, "nodePort", nodePort)

		// Get nodes for this service
		var nodes *corev1.NodeList
		if svcSpec.FilterByEndpoints {
			logger.Info("service-based node filtering enabled for service", "service", svcSpec.Name)
			filteredNodes, err := getNodesForService(ctx, clnt, &svcSpec)
			if err != nil {
				return fmt.Errorf("failed to get nodes for service %s/%s: %w", svcSpec.Namespace, svcSpec.Name, err)
			}
			nodes = filteredNodes
			logger.Info("filtered nodes for service", "service", svcSpec.Name, "nodeCount", len(nodes.Items))
		} else {
			// Use all nodes
			nodes = &corev1.NodeList{}
			if err := clnt.List(ctx, nodes); err != nil {
				return fmt.Errorf("failed to list all nodes: %w", err)
			}
			logger.Info("using all nodes for service", "service", svcSpec.Name, "nodeCount", len(nodes.Items))
		}

		// Register backends for this service (protected by LoadBalancer mutex)
		logger.Info("registering backends for service", "service", svcSpec.Name, "backendSet", serviceSpec.BackendSetName)
		if err := oci.RegisterBackends(ctx, provider, serviceSpec, nodes); err != nil {
			return fmt.Errorf("failed to register backends for service %s/%s: %w", svcSpec.Namespace, svcSpec.Name, err)
		}
		logger.Info("successfully registered backends for service", "service", svcSpec.Name)
	}

	logger.Info("successfully registered all services", "serviceCount", len(registrar.Spec.Services))
	return nil
}
