package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) error {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Getting backend set", "provider", provider)

	lbClient, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return err
	}

	request := loadbalancer.GetBackendSetRequest{
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	response, err := lbClient.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return err
	}

	logger.Info("Got Backend Set", "BackendSet", response.BackendSet)

	return nil
}
