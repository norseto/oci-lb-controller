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

package networkloadbalancer

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

func loadBalancerClient(ctx context.Context, provider common.ConfigurationProvider) (*ocilb.NetworkLoadBalancerClient, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Creating Load Network Load Balancer client", "provider", provider)
	lbClient, err := ocilb.NewNetworkLoadBalancerClientWithConfigurationProvider(provider)
	if err != nil {
		logger.Error(err, "Error creating Network Load Balancer client")
		return nil, fmt.Errorf("error creating Network Load Balancer client: %w", err)
	}
	return &lbClient, nil
}

func currentBackendSet(ctx context.Context, clnt *ocilb.NetworkLoadBalancerClient, spec api.LBRegistrarSpec) (*ocilb.GetBackendSetResponse, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "nlb", spec.LoadBalancerId)

	request := ocilb.GetBackendSetRequest{
		NetworkLoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName:        common.String(spec.BackendSetName),
	}

	response, err := clnt.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "error getting backend set")
		return nil, err
	}
	return &response, nil
}

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "nlb", spec.LoadBalancerId)
	logger.V(1).Info("Getting backend set", "provider", provider)
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

	logger := log.FromContext(ctx, "backendset", spec.BackendSetName, "nlb", spec.LoadBalancerId)
	logger.V(1).Info("Registering backend set", "provider", provider)

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return err
	}

	current, err := currentBackendSet(ctx, client, spec)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return err
	}

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
		RequestData:       currentChecker.RequestData,
		ResponseData:      currentChecker.ResponseData,
	}

	details := make([]ocilb.BackendDetails, 0)
	for _, target := range targets.Items {
		ipaddr := models.GetIPAddress(&target)
		details = append(details, ocilb.BackendDetails{
			IpAddress: &ipaddr,
			Port:      common.Int(spec.Port),
			Weight:    common.Int(spec.Weight),
		})
	}
	if len(details) < 1 {
		return fmt.Errorf("no backends found")
	}

	currentPolicy := string(current.BackendSet.Policy)
	request := ocilb.UpdateBackendSetRequest{
		UpdateBackendSetDetails: ocilb.UpdateBackendSetDetails{
			Backends:                                details,
			HealthChecker:                           &healthChecker,
			Policy:                                  common.String(currentPolicy),
			IsPreserveSource:                        current.IsPreserveSource,
			IsFailOpen:                              current.IsFailOpen,
			IsInstantFailoverEnabled:                current.IsInstantFailoverEnabled,
			IpVersion:                               current.IpVersion,
			IsInstantFailoverTcpResetEnabled:        current.IsInstantFailoverTcpResetEnabled,
			AreOperationallyActiveBackendsPreferred: current.AreOperationallyActiveBackendsPreferred,
		},
		NetworkLoadBalancerId: common.String(spec.LoadBalancerId),
		BackendSetName:        common.String(spec.BackendSetName),
	}

	response, err := client.UpdateBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error updating backend set", "response", response, "request", request)
		return fmt.Errorf("error updating backend set: %w", err)
	}

	logger.V(2).Info("Updated Backend Set")

	return nil
}
