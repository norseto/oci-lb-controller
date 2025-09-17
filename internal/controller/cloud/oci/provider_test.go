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
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/onsi/gomega"
	"github.com/oracle/oci-go-sdk/v65/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
)

func TestNewConfigurationProvider(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := &api.ApiKeySpec{
		User:        "test-user",
		Fingerprint: "test-fingerprint",
		Tenancy:     "test-tenancy",
		Region:      "us-ashburn-1",
		PrivateKey: api.PrivateKeySpec{
			Namespace: "default",
			SecretKeyRef: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-secret",
				},
				Key: "private-key",
			},
		},
	}
	privateKey := "test-private-key-content"

	provider, err := NewConfigurationProvider(ctx, spec, privateKey)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(provider).ToNot(gomega.BeNil())

	// Test that the provider implements the ConfigurationProvider interface
	_, ok := provider.(common.ConfigurationProvider)
	g.Expect(ok).To(gomega.BeTrue())
}

func TestNewConfigurationProvider_WithEmptySpec(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := &api.ApiKeySpec{}
	privateKey := "test-private-key-content"

	provider, err := NewConfigurationProvider(ctx, spec, privateKey)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(provider).ToNot(gomega.BeNil())
}

func TestIsNetworkLoadBalancer(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test with Network Load Balancer ID
	spec1 := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.networkloadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	result1 := isNetworkLoadBalancer(spec1)
	g.Expect(result1).To(gomega.BeTrue())

	// Test with Application Load Balancer ID
	spec2 := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.loadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	result2 := isNetworkLoadBalancer(spec2)
	g.Expect(result2).To(gomega.BeFalse())

	// Test with empty Load Balancer ID
	spec3 := api.LBRegistrarSpec{
		LoadBalancerId: "",
	}
	result3 := isNetworkLoadBalancer(spec3)
	g.Expect(result3).To(gomega.BeFalse())

	// Test with partial match (should still match)
	spec4 := api.LBRegistrarSpec{
		LoadBalancerId: "some-id.networkloadbalancer.oc1.iad",
	}
	result4 := isNetworkLoadBalancer(spec4)
	g.Expect(result4).To(gomega.BeTrue())
}

func TestGetBackendSet_NetworkLoadBalancer(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.networkloadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BackendSetName: "test-backend-set",
	}

	// Mock provider that will cause the underlying functions to fail
	provider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	targets, err := GetBackendSet(ctx, provider, spec)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(targets).To(gomega.BeNil())
}

func TestGetBackendSet_ApplicationLoadBalancer(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.loadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BackendSetName: "test-backend-set",
	}

	// Mock provider that will cause the underlying functions to fail
	provider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	targets, err := GetBackendSet(ctx, provider, spec)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(targets).To(gomega.BeNil())
}

func TestRegisterBackends_NetworkLoadBalancer(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.networkloadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BackendSetName: "test-backend-set",
		NodePort:       30080,
		Weight:         1,
	}

	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			},
		},
	}

	// Mock provider that will cause the underlying functions to fail
	provider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	err := RegisterBackends(ctx, provider, spec, nodes)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestRegisterBackends_ApplicationLoadBalancer(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "ocid1.loadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		BackendSetName: "test-backend-set",
		NodePort:       30080,
		Weight:         1,
	}

	nodes := &corev1.NodeList{
		Items: []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			},
		},
	}

	// Mock provider that will cause the underlying functions to fail
	provider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	err := RegisterBackends(ctx, provider, spec, nodes)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestLoadBalancerTypeDetection(t *testing.T) {
	g := gomega.NewWithT(t)

	testCases := []struct {
		name           string
		loadBalancerId string
		expected       bool
	}{
		{
			name:           "Network Load Balancer with full OCID",
			loadBalancerId: "ocid1.networkloadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:       true,
		},
		{
			name:           "Application Load Balancer with full OCID",
			loadBalancerId: "ocid1.loadbalancer.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:       false,
		},
		{
			name:           "Network Load Balancer with partial OCID",
			loadBalancerId: "some-id.networkloadbalancer.oc1.iad",
			expected:       true,
		},
		{
			name:           "Application Load Balancer with partial OCID",
			loadBalancerId: "some-id.loadbalancer.oc1.iad",
			expected:       false,
		},
		{
			name:           "Empty Load Balancer ID",
			loadBalancerId: "",
			expected:       false,
		},
		{
			name:           "Invalid Load Balancer ID",
			loadBalancerId: "invalid-id",
			expected:       false,
		},
		{
			name:           "Network Load Balancer in different region",
			loadBalancerId: "ocid1.networkloadbalancer.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:       true,
		},
		{
			name:           "Application Load Balancer in different region",
			loadBalancerId: "ocid1.loadbalancer.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := api.LBRegistrarSpec{
				LoadBalancerId: tc.loadBalancerId,
			}
			result := isNetworkLoadBalancer(spec)
			g.Expect(result).To(gomega.Equal(tc.expected))
		})
	}
}

func TestConfigurationProvider_Interface(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := &api.ApiKeySpec{
		User:        "test-user",
		Fingerprint: "test-fingerprint",
		Tenancy:     "test-tenancy",
		Region:      "us-ashburn-1",
		PrivateKey: api.PrivateKeySpec{
			Namespace: "default",
			SecretKeyRef: corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-secret",
				},
				Key: "private-key",
			},
		},
	}
	privateKey := "test-private-key-content"

	provider, err := NewConfigurationProvider(ctx, spec, privateKey)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(provider).ToNot(gomega.BeNil())

	// Test that the provider implements all required methods
	_, ok := provider.(common.ConfigurationProvider)
	g.Expect(ok).To(gomega.BeTrue())

	// Test that we can call the methods (they will return the configured values)
	user, err := provider.UserOCID()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(user).To(gomega.Equal("test-user"))

	tenancy, err := provider.TenancyOCID()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(tenancy).To(gomega.Equal("test-tenancy"))

	region, err := provider.Region()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(region).To(gomega.Equal("us-ashburn-1"))

	fingerprint, err := provider.KeyFingerprint()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(fingerprint).To(gomega.Equal("test-fingerprint"))

	keyID, err := provider.KeyID()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(keyID).ToNot(gomega.BeEmpty())

	privateRSAKey, err := provider.PrivateRSAKey()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(privateRSAKey).To(gomega.Equal("test-private-key-content"))
}

// MockConfigurationProvider implements common.ConfigurationProvider for testing
type MockConfigurationProvider struct {
	Error error
}

func (m *MockConfigurationProvider) UserOCID() (string, error) {
	return "test-user-ocid", m.Error
}

func (m *MockConfigurationProvider) TenancyOCID() (string, error) {
	return "test-tenancy-ocid", m.Error
}

func (m *MockConfigurationProvider) KeyFingerprint() (string, error) {
	return "test-fingerprint", m.Error
}

func (m *MockConfigurationProvider) Region() (string, error) {
	return "us-ashburn-1", m.Error
}

func (m *MockConfigurationProvider) KeyID() (string, error) {
	return "test-key-id", m.Error
}

func (m *MockConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return nil, m.Error
}

func (m *MockConfigurationProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{}, m.Error
}
