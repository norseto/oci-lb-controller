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
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	corev1 "k8s.io/api/core/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

func TestIsNetworkLoadBalancer(t *testing.T) {
	var testCases = []struct {
		name     string
		spec     api.LBRegistrarSpec
		expected bool
	}{
		{
			"Network Load Balancer",
			api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1.phx.exampleuniqueID"},
			true,
		},
		{
			"Load Balancer",
			api.LBRegistrarSpec{LoadBalancerId: "ocid1.loadbalancer.oc1.phx.exampleuniqueID"},
			false,
		},
		{"Empty Load Balancer", api.LBRegistrarSpec{LoadBalancerId: ""}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isNetworkLoadBalancer(tc.spec)
			if actual != tc.expected {
				t.Errorf("Unexpected result: expected %v, actual %v", tc.expected, actual)
			}
		})
	}
}

func TestGetBackendSetDelegation(t *testing.T) {
	origLB := loadBalancerGetBackendSet
	origNLB := networkLoadBalancerGetBackendSet
	defer func() {
		loadBalancerGetBackendSet = origLB
		networkLoadBalancerGetBackendSet = origNLB
	}()

	calledLB := false
	loadBalancerGetBackendSet = func(
		context.Context,
		common.ConfigurationProvider,
		api.LBRegistrarSpec,
	) ([]*models.LoadBalanceTarget, error) {
		calledLB = true
		return []*models.LoadBalanceTarget{{Name: "lb"}}, nil
	}

	calledNLB := false
	networkLoadBalancerGetBackendSet = func(
		context.Context,
		common.ConfigurationProvider,
		api.LBRegistrarSpec,
	) ([]*models.LoadBalanceTarget, error) {
		calledNLB = true
		return []*models.LoadBalanceTarget{{Name: "nlb"}}, nil
	}

	targets, err := GetBackendSet(
		context.Background(),
		nil,
		api.LBRegistrarSpec{LoadBalancerId: "ocid1.loadbalancer"},
	)
	if err != nil || !calledLB || calledNLB || targets[0].Name != "lb" {
		t.Fatalf(
			"expected load balancer path: calledLB=%v calledNLB=%v targets=%v err=%v",
			calledLB,
			calledNLB,
			targets,
			err,
		)
	}

	calledLB = false
	calledNLB = false
	targets, err = GetBackendSet(
		context.Background(),
		nil,
		api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1"},
	)
	if err != nil || !calledNLB || calledLB || targets[0].Name != "nlb" {
		t.Fatalf(
			"expected network load balancer path: calledLB=%v calledNLB=%v targets=%v err=%v",
			calledLB,
			calledNLB,
			targets,
			err,
		)
	}
}

func TestRegisterBackendsDelegation(t *testing.T) {
	origLB := loadBalancerRegisterBackends
	origNLB := networkRegisterBackends
	defer func() {
		loadBalancerRegisterBackends = origLB
		networkRegisterBackends = origNLB
	}()

	calledLB := false
	loadBalancerRegisterBackends = func(
		context.Context,
		common.ConfigurationProvider,
		api.LBRegistrarSpec,
		*corev1.NodeList,
	) error {
		calledLB = true
		return nil
	}

	calledNLB := false
	networkRegisterBackends = func(
		context.Context,
		common.ConfigurationProvider,
		api.LBRegistrarSpec,
		*corev1.NodeList,
	) error {
		calledNLB = true
		return nil
	}

	if err := RegisterBackends(
		context.Background(),
		nil,
		api.LBRegistrarSpec{LoadBalancerId: "ocid1.loadbalancer"},
		&corev1.NodeList{},
	); err != nil || !calledLB || calledNLB {
		t.Fatalf(
			"expected LB backend registration path, err=%v calledLB=%v calledNLB=%v",
			err,
			calledLB,
			calledNLB,
		)
	}

	calledLB = false
	calledNLB = false
	if err := RegisterBackends(
		context.Background(),
		nil,
		api.LBRegistrarSpec{LoadBalancerId: "ocid1.networkloadbalancer.oc1"},
		&corev1.NodeList{},
	); err != nil || !calledNLB || calledLB {
		t.Fatalf(
			"expected NLB backend registration path, err=%v calledLB=%v calledNLB=%v",
			err,
			calledLB,
			calledNLB,
		)
	}
}

func TestNewConfigurationProvider(t *testing.T) {
	ctx := context.Background()
	spec := &api.ApiKeySpec{
		Tenancy:     "tenancy",
		User:        "user",
		Region:      "region",
		Fingerprint: "fingerprint",
	}
	provider, err := NewConfigurationProvider(ctx, spec, "private-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tenancy, err := provider.TenancyOCID()
	if err != nil || tenancy != "tenancy" {
		t.Fatalf("bad tenancy %v %v", tenancy, err)
	}
	user, _ := provider.UserOCID()
	if user != "user" {
		t.Fatalf("bad user %s", user)
	}
}
