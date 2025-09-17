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

package loadbalancer

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/loadbalancer"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

// LoadBalancerClient abstracts the OCI load balancer client.
type LoadBalancerClient interface {
	GetBackendSet(context.Context, ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error)
	UpdateBackendSet(context.Context, ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error)
}

type ociLBClient struct {
	*ocilb.LoadBalancerClient
}

func (c *ociLBClient) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	return c.LoadBalancerClient.GetBackendSet(ctx, req)
}

func (c *ociLBClient) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
	return c.LoadBalancerClient.UpdateBackendSet(ctx, req)
}

var newLBClient = func(provider common.ConfigurationProvider) (LoadBalancerClient, error) {
	lbClient, err := ocilb.NewLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}
	return &ociLBClient{&lbClient}, nil
}

func loadBalancerClient(ctx context.Context, provider common.ConfigurationProvider) (LoadBalancerClient, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Creating Load Balancer client", "provider", provider)
	clnt, err := newLBClient(provider)
	if err != nil {
		logger.Error(err, "Error creating Load Balancer client")
		return nil, fmt.Errorf("Error creating Load Balancer client: %w", err)
	}
	return clnt, nil
}

func currentBackendSet(ctx context.Context, clnt LoadBalancerClient, spec api.LBRegistrarSpec) (*ocilb.GetBackendSetResponse, error) {
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
	logger.V(1).Info("Getting backend set", "provider", provider)
	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return nil, err
	}

	response, err := currentBackendSet(ctx, client, spec)
	if err != nil {
		logger.Error(err, "error getting backend set")
		return nil, err
	}

	logger.V(2).Info("got Backend Set", "BackendSet", response.BackendSet)
	targets := make([]*models.LoadBalanceTarget, 0, len(response.BackendSet.Backends))
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
	logger.V(1).Info("registering backend set", "provider", provider)

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return err
	}

	current, err := currentBackendSet(ctx, client, spec)
	if err != nil {
		logger.Error(err, "error getting backend set")
		return err
	}

	port := spec.NodePort
	if spec.Port != 0 {
		port = spec.Port
	}
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
		return fmt.Errorf("Error updating backend set: %w", err)
	}

	logger.V(2).Info("Updated Backend Set")

	return nil
}
