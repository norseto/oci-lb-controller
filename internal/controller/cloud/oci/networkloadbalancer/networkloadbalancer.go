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
	"time"

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
	client, err := loadBalancerClient(ctx, provider)
	if err != nil {
		return nil, err
	}

	response, err := currentBackendSet(ctx, client, spec)
	if err != nil {
		logger.Error(err, "Error getting backend set")
		return nil, err
	}

	logger.V(2).Info("Got Backend Set", "BackendSet", response.BackendSet)
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

func determinePort(spec api.LBRegistrarSpec) int {
	if spec.Port != 0 {
		return spec.Port
	}
	return spec.NodePort
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
			Port:      common.Int(determinePort(spec)),
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

	logger.V(2).Info("Updated Backend Set, waiting for WorkRequest completion", "workRequestId", response.OpcWorkRequestId)

	// Wait for WorkRequest completion to avoid conflicts
	if err := waitForWorkRequestCompletion(ctx, client, response.OpcWorkRequestId); err != nil {
		logger.Error(err, "WorkRequest failed", "workRequestId", response.OpcWorkRequestId)
		return fmt.Errorf("work request failed: %w", err)
	}

	logger.Info("WorkRequest completed successfully", "workRequestId", response.OpcWorkRequestId)

	return nil
}

// waitForWorkRequestCompletion waits for a WorkRequest to complete
func waitForWorkRequestCompletion(ctx context.Context, client *ocilb.NetworkLoadBalancerClient, workRequestId *string) error {
	logger := log.FromContext(ctx)

	if workRequestId == nil {
		logger.V(1).Info("No WorkRequest ID provided, skipping wait")
		return nil
	}

	logger.Info("Waiting for WorkRequest completion", "workRequestId", *workRequestId)

	maxAttempts := 60 // 5 minutes max (60 * 5 seconds)
	attempt := 0

	for attempt < maxAttempts {
		workReq, err := client.GetWorkRequest(ctx, ocilb.GetWorkRequestRequest{
			WorkRequestId: workRequestId,
		})
		if err != nil {
			logger.Error(err, "Error getting WorkRequest status", "workRequestId", *workRequestId)
			return fmt.Errorf("error getting work request status: %w", err)
		}

		logger.V(1).Info("WorkRequest status check", "workRequestId", *workRequestId, "status", workReq.Status, "attempt", attempt)

		switch workReq.Status {
		case ocilb.OperationStatusSucceeded:
			logger.Info("WorkRequest completed successfully", "workRequestId", *workRequestId)
			return nil
		case ocilb.OperationStatusFailed:
			logger.Error(nil, "WorkRequest failed", "workRequestId", *workRequestId)
			return fmt.Errorf("work request failed: %s", *workRequestId)
		case ocilb.OperationStatusCanceled:
			logger.Error(nil, "WorkRequest was canceled", "workRequestId", *workRequestId)
			return fmt.Errorf("work request canceled: %s", *workRequestId)
		case ocilb.OperationStatusInProgress, ocilb.OperationStatusAccepted:
			// Continue waiting
			logger.V(1).Info("WorkRequest still in progress, waiting...", "workRequestId", *workRequestId, "status", workReq.Status)
		default:
			logger.V(1).Info("WorkRequest in unknown state, continuing to wait", "workRequestId", *workRequestId, "status", workReq.Status)
		}

		attempt++
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				// Continue to next iteration
			}
		}
	}

	return fmt.Errorf("timeout waiting for work request completion: %s", *workRequestId)
}
