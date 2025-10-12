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

package oci

import (
	"context"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	alb "github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/loadbalancer"
	nlb "github.com/norseto/oci-lb-controller/internal/controller/cloud/oci/networkloadbalancer"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

var (
	loadBalancerGetBackendSet        = alb.GetBackendSet
	networkLoadBalancerGetBackendSet = nlb.GetBackendSet
	loadBalancerRegisterBackends     = alb.RegisterBackends
	networkRegisterBackends          = nlb.RegisterBackends
)

// NewConfigurationProvider is a function that creates a new instance of the ConfigurationProvider interface.
// It takes in a context.Context object, a pointer to an api.ApiKeySpec object
func NewConfigurationProvider(ctx context.Context, spec *api.ApiKeySpec, privateKey string) (common.ConfigurationProvider, error) {
	_ = log.FromContext(ctx)

	key := api.ApiKeySpec{}
	var pass string
	spec.DeepCopyInto(&key)

	provider := common.NewRawConfigurationProvider(
		key.Tenancy, key.User, key.Region, key.Fingerprint, privateKey, &pass)
	return provider, nil
}

func isNetworkLoadBalancer(spec api.LBRegistrarSpec) bool {
	return strings.Contains(spec.LoadBalancerId, ".networkloadbalancer.")
}

func GetBackendSet(ctx context.Context, provider common.ConfigurationProvider, spec api.LBRegistrarSpec) ([]*models.LoadBalanceTarget, error) {
	if isNetworkLoadBalancer(spec) {
		return networkLoadBalancerGetBackendSet(ctx, provider, spec)
	}
	return loadBalancerGetBackendSet(ctx, provider, spec)
}

func RegisterBackends(ctx context.Context, provider common.ConfigurationProvider,
	spec api.LBRegistrarSpec, targets *corev1.NodeList) error {
	if isNetworkLoadBalancer(spec) {
		return networkRegisterBackends(ctx, provider, spec, targets)
	}
	return loadBalancerRegisterBackends(ctx, provider, spec, targets)
}
