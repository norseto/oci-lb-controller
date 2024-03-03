package oci

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Getting backend set", "provider", provider)
	var targets []*models.LoadBalanceTarget

	lbClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return targets, errors.Wrap(err, "Error creating Load Balancer client")
	}

	request := loadbalancer.GetBackendSetRequest{
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	response, err := lbClient.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return targets, errors.Wrap(err, "Error getting backend set")
	}

	logger.V(2).Info("Got Backend Set", "BackendSet", response.BackendSet)
	for _, backend := range response.BackendSet.Backends {
		targets = append(targets, &models.LoadBalanceTarget{
			Name:      *backend.Name,
			IpAddress: *backend.IpAddress,
			Port:      *backend.Port,
			Weight:    *backend.Weight,
		})
	}

	return targets, nil
}

func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider,
	spec api.LBRegistrarSpec, targets *corev1.NodeList) error {

	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Registering backend set", "provider", provider)

	lbClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return errors.Wrap(err, "Error creating Load Balancer client")
	}

	port := spec.Port
	weight := spec.Weight

	details := make([]loadbalancer.BackendDetails, 0)
	for _, target := range targets.Items {
		ipaddr := models.GetIPAddress(&target)
		details = append(details, loadbalancer.BackendDetails{
			IpAddress: &ipaddr,
			Port:      &port,
			Weight:    &weight,
		})
	}

	request := loadbalancer.UpdateBackendSetRequest{
		UpdateBackendSetDetails: loadbalancer.UpdateBackendSetDetails{
			Backends: details,
		},
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	_, err = lbClient.UpdateBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error updating backend set")
		return errors.Wrap(err, "Error getting backend set")
	}

	logger.V(2).Info("Updated Backend Set")

	return nil
}
