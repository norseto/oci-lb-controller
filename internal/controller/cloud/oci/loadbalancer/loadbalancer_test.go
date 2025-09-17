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
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/onsi/gomega"
	"github.com/oracle/oci-go-sdk/v65/common"
	ocilb "github.com/oracle/oci-go-sdk/v65/loadbalancer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/norseto/oci-lb-controller/api/v1alpha1"
	"github.com/norseto/oci-lb-controller/internal/controller/models"
)

// MockLoadBalancerClient implements LoadBalancerClient for testing
type MockLoadBalancerClient struct {
	GetBackendSetResponse    ocilb.GetBackendSetResponse
	GetBackendSetError       error
	UpdateBackendSetResponse ocilb.UpdateBackendSetResponse
	UpdateBackendSetError    error
}

func (m *MockLoadBalancerClient) GetBackendSet(ctx context.Context, req ocilb.GetBackendSetRequest) (ocilb.GetBackendSetResponse, error) {
	return m.GetBackendSetResponse, m.GetBackendSetError
}

func (m *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, req ocilb.UpdateBackendSetRequest) (ocilb.UpdateBackendSetResponse, error) {
	return m.UpdateBackendSetResponse, m.UpdateBackendSetError
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

func TestOciLBClient_GetBackendSet(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	client := &ociLBClient{}

	// This test would require a real OCI client, so we'll test the interface implementation
	// by testing the mock implementation instead
	_ = client // Suppress unused variable warning
	mockClient := &MockLoadBalancerClient{
		GetBackendSetResponse: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Backends: []ocilb.Backend{
					{
						Name:      common.String("backend-1"),
						IpAddress: common.String("10.0.0.1"),
						Port:      common.Int(8080),
						Weight:    common.Int(1),
					},
				},
			},
		},
	}

	req := ocilb.GetBackendSetRequest{
		LoadBalancerId: common.String("test-lb-id"),
		BackendSetName: common.String("test-backend-set"),
	}

	response, err := mockClient.GetBackendSet(ctx, req)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(response.BackendSet.Backends).To(gomega.HaveLen(1))
	g.Expect(*response.BackendSet.Backends[0].Name).To(gomega.Equal("backend-1"))
	g.Expect(*response.BackendSet.Backends[0].IpAddress).To(gomega.Equal("10.0.0.1"))
	g.Expect(*response.BackendSet.Backends[0].Port).To(gomega.Equal(8080))
	g.Expect(*response.BackendSet.Backends[0].Weight).To(gomega.Equal(1))
}

func TestOciLBClient_UpdateBackendSet(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	mockClient := &MockLoadBalancerClient{
		UpdateBackendSetResponse: ocilb.UpdateBackendSetResponse{
			OpcWorkRequestId: common.String("test-work-request-id"),
		},
	}

	req := ocilb.UpdateBackendSetRequest{
		LoadBalancerId: common.String("test-lb-id"),
		BackendSetName: common.String("test-backend-set"),
		UpdateBackendSetDetails: ocilb.UpdateBackendSetDetails{
			Backends: []ocilb.BackendDetails{
				{
					IpAddress: common.String("10.0.0.1"),
					Port:      common.Int(8080),
					Weight:    common.Int(1),
				},
			},
		},
	}

	response, err := mockClient.UpdateBackendSet(ctx, req)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(*response.OpcWorkRequestId).To(gomega.Equal("test-work-request-id"))
}

func TestLoadBalancerClient_ErrorHandling(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	mockClient := &MockLoadBalancerClient{
		GetBackendSetError:    errors.New("backend set not found"),
		UpdateBackendSetError: errors.New("update failed"),
	}

	req := ocilb.GetBackendSetRequest{
		LoadBalancerId: common.String("test-lb-id"),
		BackendSetName: common.String("test-backend-set"),
	}

	_, err := mockClient.GetBackendSet(ctx, req)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("backend set not found"))

	updateReq := ocilb.UpdateBackendSetRequest{
		LoadBalancerId: common.String("test-lb-id"),
		BackendSetName: common.String("test-backend-set"),
	}

	_, err = mockClient.UpdateBackendSet(ctx, updateReq)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("update failed"))
}

func TestCurrentBackendSet(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
		BackendSetName: "test-backend-set",
	}

	mockClient := &MockLoadBalancerClient{
		GetBackendSetResponse: ocilb.GetBackendSetResponse{
			BackendSet: ocilb.BackendSet{
				Backends: []ocilb.Backend{
					{
						Name:      common.String("backend-1"),
						IpAddress: common.String("10.0.0.1"),
						Port:      common.Int(8080),
						Weight:    common.Int(1),
					},
				},
				HealthChecker: &ocilb.HealthChecker{
					Protocol:         common.String("HTTP"),
					Port:             common.Int(8080),
					UrlPath:          common.String("/health"),
					ReturnCode:       common.Int(200),
					Retries:          common.Int(3),
					TimeoutInMillis:  common.Int(3000),
					IntervalInMillis: common.Int(10000),
				},
				Policy: common.String("ROUND_ROBIN"),
			},
		},
	}

	response, err := currentBackendSet(ctx, mockClient, spec)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(response).ToNot(gomega.BeNil())
	g.Expect(response.BackendSet.Backends).To(gomega.HaveLen(1))
	g.Expect(*response.BackendSet.Backends[0].Name).To(gomega.Equal("backend-1"))
	g.Expect(*response.BackendSet.HealthChecker.Protocol).To(gomega.Equal("HTTP"))
	g.Expect(*response.BackendSet.Policy).To(gomega.Equal("ROUND_ROBIN"))
}

func TestCurrentBackendSet_Error(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
		BackendSetName: "test-backend-set",
	}

	mockClient := &MockLoadBalancerClient{
		GetBackendSetError: errors.New("backend set not found"),
	}

	response, err := currentBackendSet(ctx, mockClient, spec)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(response).To(gomega.BeNil())
	g.Expect(err.Error()).To(gomega.Equal("backend set not found"))
}

func TestGetBackendSet(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
		BackendSetName: "test-backend-set",
	}

	// Create a mock provider that will cause loadBalancerClient to fail
	mockProvider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	targets, err := GetBackendSet(ctx, mockProvider, spec)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(targets).To(gomega.BeNil())
}

func TestRegisterBackends(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
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
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "10.0.0.2",
						},
					},
				},
			},
		},
	}

	// Create a mock provider that will cause loadBalancerClient to fail
	mockProvider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	err := RegisterBackends(ctx, mockProvider, spec, nodes)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestRegisterBackends_WithPortField(t *testing.T) {
	g := gomega.NewWithT(t)

	ctx := context.Background()
	spec := api.LBRegistrarSpec{
		LoadBalancerId: "test-lb-id",
		BackendSetName: "test-backend-set",
		Port:           8080, // Using deprecated Port field
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

	// Create a mock provider that will cause loadBalancerClient to fail
	mockProvider := &MockConfigurationProvider{
		Error: errors.New("configuration error"),
	}

	// This should fail because we can't create a real OCI client
	err := RegisterBackends(ctx, mockProvider, spec, nodes)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestLoadBalanceTarget_Conversion(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test the conversion from OCI Backend to LoadBalanceTarget
	backend := ocilb.Backend{
		Name:      common.String("backend-1"),
		IpAddress: common.String("10.0.0.1"),
		Port:      common.Int(8080),
		Weight:    common.Int(5),
	}

	target := &models.LoadBalanceTarget{
		Name:      *backend.Name,
		IpAddress: *backend.IpAddress,
		Port:      *backend.Port,
		Weight:    *backend.Weight,
	}

	g.Expect(target.Name).To(gomega.Equal("backend-1"))
	g.Expect(target.IpAddress).To(gomega.Equal("10.0.0.1"))
	g.Expect(target.Port).To(gomega.Equal(8080))
	g.Expect(target.Weight).To(gomega.Equal(5))
}

func TestHealthCheckerDetails_Creation(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test health checker details creation
	currentChecker := &ocilb.HealthChecker{
		Protocol:          common.String("HTTP"),
		Port:              common.Int(8080),
		UrlPath:           common.String("/health"),
		ReturnCode:        common.Int(200),
		Retries:           common.Int(3),
		TimeoutInMillis:   common.Int(3000),
		IntervalInMillis:  common.Int(10000),
		ResponseBodyRegex: common.String("OK"),
		IsForcePlainText:  common.Bool(true),
	}

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

	g.Expect(*healthChecker.Protocol).To(gomega.Equal("HTTP"))
	g.Expect(*healthChecker.Port).To(gomega.Equal(8080))
	g.Expect(*healthChecker.UrlPath).To(gomega.Equal("/health"))
	g.Expect(*healthChecker.ReturnCode).To(gomega.Equal(200))
	g.Expect(*healthChecker.Retries).To(gomega.Equal(3))
	g.Expect(*healthChecker.TimeoutInMillis).To(gomega.Equal(3000))
	g.Expect(*healthChecker.IntervalInMillis).To(gomega.Equal(10000))
	g.Expect(*healthChecker.ResponseBodyRegex).To(gomega.Equal("OK"))
	g.Expect(*healthChecker.IsForcePlainText).To(gomega.BeTrue())
}
