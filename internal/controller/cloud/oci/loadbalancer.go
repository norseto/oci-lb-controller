package oci

import (
	"context"
	"github.com/norseto/oci-lb-controller/internal/controller/models"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Getting backend set", "provider", provider)
	var targets []*models.LoadBalanceTarget

	lbClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return targets, err
	}

	request := loadbalancer.GetBackendSetRequest{
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	response, err := lbClient.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return targets, err
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
