package oci

import (
	"context"

	"github.com/pkg/errors"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/loadbalancer"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

func loadBalancerClient(ctx context.Context, provider common.ConfigurationProvider) (*ocilb.LoadBalancerClient, error) {
	logger := log.FromContext(ctx)
	logger.Info("Creating Load Balancer client", "provider", provider)
	lbClient, err := ocilb.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return nil, errors.Wrap(err, "Error creating Load Balancer client")
	}
	return &lbClient, nil
}

func currentBackendSet(ctx context.Context, clnt *ocilb.LoadBalancerClient, spec api.LBRegistrarSpec) (*ocilb.GetBackendSetResponse, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)

	request := ocilb.GetBackendSetRequest{
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	response, err := clnt.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return nil, err
	}
	return &response, nil
}

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Getting backend set", "provider", provider)
	var targets []*models.LoadBalanceTarget

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return targets, err
	}

	response, err := currentBackendSet(ctx, client, spec)
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

func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider,
	spec api.LBRegistrarSpec, targets *corev1.NodeList) error {

	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "lb", spec.LoadBalancerId)
	logger.Info("Registering backend set", "provider", provider)

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return err
	}

	current, err := currentBackendSet(ctx, client, spec)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return err
	}

	port := spec.Port
	weight := spec.Weight
	currentChecker := current.BackendSet.HealthChecker
	healthChecker := ocilb.HealthCheckerDetails{
		Protocol:          currentChecker.Protocol,
		Port:              currentChecker.Port,
		UrlPath:           currentChecker.UrlPath,
		ReturnCode:        currentChecker.ReturnCode,
		Retries:           currentChecker.Retries,
		TimeoutInMillis:   currentChecker.TimeoutInMillis,
		IntervalInMillis:  currentChecker.IntervalInMillis,
		ResponseBodyRegex: currentChecker.ResponseBodyRegex,
		IsForcePlainText:  currentChecker.IsForcePlainText,
	}

	details := make([]ocilb.BackendDetails, 0)
	for _, target := range targets.Items {
		ipaddr := models.GetIPAddress(&target)
		details = append(details, ocilb.BackendDetails{
			IpAddress: &ipaddr,
			Port:      &port,
			Weight:    &weight,
		})
	}

	request := ocilb.UpdateBackendSetRequest{
		UpdateBackendSetDetails: ocilb.UpdateBackendSetDetails{
			Backends:      details,
			HealthChecker: &healthChecker,
			Policy:        current.BackendSet.Policy,
		},
		LoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName: common.String(spec.BackendSetName),
	}

	_, err = client.UpdateBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error updating backend set")
		return errors.Wrap(err, "Error getting backend set")
	}

	logger.V(2).Info("Updated Backend Set")

	return nil
}
