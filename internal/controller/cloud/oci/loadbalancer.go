/*
MIT License

Copyright (c) 2024 Norihiro Seto

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package oci

import (
	"context"

	"github.com/pkg/errors"

	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

func currentBackendSet(ctx context.Context, clnt *ocilb.LoadBalancerClient, tg models.TargetGroup) (*ocilb.GetBackendSetResponse, error) {
	logger := log.FromContext(ctx, "backendset", tg.Name, "lb", tg.LoadBalancerId)

	request := ocilb.GetBackendSetRequest{
		LoadBalancerId: common.String(tg.LoadBalancerId),
		BackendSetName: common.String(tg.Name),
	}

	response, err := clnt.GetBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return nil, err
	}
	return &response, nil
}

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, tg models.TargetGroup) ([]*models.LoadBalanceTarget, error) {
	logger := log.FromContext(ctx, "backendset", tg.Name, "lb", tg.LoadBalancerId)
	logger.Info("Getting backend set", "provider", provider)
	var targets []*models.LoadBalanceTarget

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return targets, err
	}

	response, err := currentBackendSet(ctx, client, tg)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return targets, err
	}

	logger.V(2).Info("Got Backend Set", "BackendSet", response.BackendSet)
	for _, backend := range response.BackendSet.Backends {
		targets = append(targets, &models.LoadBalanceTarget{
			IpAddress: *backend.IpAddress,
			Port:      *backend.Port,
			Weight:    *backend.Weight,
		})
	}

	return targets, nil
}

func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider,
	tg models.TargetGroup, targets []*models.LoadBalanceTarget) error {

	logger := log.FromContext(ctx, "backendset", tg.Name, "lb", tg.LoadBalancerId)
	logger.Info("Registering backend set", "provider", provider)

	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return err
	}

	current, err := currentBackendSet(ctx, client, tg)
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
		IsForcePlainText:  currentChecker.IsForcePlainText,
	}

	details := make([]ocilb.BackendDetails, 0)
	for _, target := range targets {
		ipaddr := target.IpAddress
		weight := target.Weight
		port := target.Port
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
		LoadBalancerId: common.String(tg.LoadBalancerId),
		BackendSetName: common.String(tg.Name),
	}

	_, err = client.UpdateBackendSet(ctx, request)
	if err != nil {
		logger.Error(err, "Error updating backend set")
		return errors.Wrap(err, "Error getting backend set")
	}

	logger.V(2).Info("Updated Backend Set")

	return nil
}
